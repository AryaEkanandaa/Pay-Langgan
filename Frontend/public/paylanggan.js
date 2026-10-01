(function () {
  'use strict'

  var script = document.currentScript
  if (!script || !script.src) return
  var checkoutOrigin = new URL(script.src).origin
  var active = null

  function close() {
    if (!active) return
    var state = active
    active = null
    window.removeEventListener('message', state.onMessage)
    document.removeEventListener('keydown', state.onKeydown)
    state.overlay.remove()
    document.body.style.overflow = state.previousOverflow
    if (state.previousFocus && state.previousFocus.focus) state.previousFocus.focus()
    if (typeof state.onClose === 'function') state.onClose()
  }

  function open(options) {
    if (!options || typeof options.checkoutUrl !== 'string') throw new Error('checkoutUrl is required')
    var url = new URL(options.checkoutUrl)
    if (url.origin !== checkoutOrigin || !/^\/checkout\/[A-Za-z0-9_-]+$/.test(url.pathname)) {
      throw new Error('Invalid PayLanggan checkout URL')
    }
    close()
    url.searchParams.set('parent_origin', window.location.origin)

    var overlay = document.createElement('div')
    overlay.className = 'pl-dropin-overlay'
    overlay.setAttribute('role', 'dialog')
    overlay.setAttribute('aria-modal', 'true')
    overlay.setAttribute('aria-label', 'Pembayaran PayLanggan')
    var panel = document.createElement('div')
    panel.className = 'pl-dropin-panel'
    var button = document.createElement('button')
    button.type = 'button'
    button.className = 'pl-dropin-close'
    button.setAttribute('aria-label', 'Tutup pembayaran')
    button.textContent = '×'
    var fullPage = document.createElement('a')
    fullPage.className = 'pl-dropin-full-page'
    fullPage.href = options.checkoutUrl
    fullPage.target = '_blank'
    fullPage.rel = 'noopener noreferrer'
    fullPage.textContent = 'Buka halaman penuh'
    var frame = document.createElement('iframe')
    frame.title = 'Pembayaran PayLanggan'
    frame.src = url.toString()
    frame.allow = 'payment *'
    var controls = document.createElement('div')
    controls.className = 'pl-dropin-controls'
    controls.appendChild(fullPage)
    controls.appendChild(button)
    panel.appendChild(controls)
    panel.appendChild(frame)
    overlay.appendChild(panel)

    function onMessage(event) {
      if (event.origin !== checkoutOrigin || event.source !== frame.contentWindow) return
      if (!event.data || event.data.type !== 'paylanggan:checkout-status' || event.data.status !== 'paid') return
      if (typeof options.onPaid === 'function') options.onPaid()
    }
    function onKeydown(event) {
      if (event.key === 'Escape') close()
      if (event.key === 'Tab') {
        // The iframe contains its own focusable checkout controls.
        if (document.activeElement === button && event.shiftKey) { event.preventDefault(); frame.focus() }
        else if (document.activeElement === frame && !event.shiftKey) { event.preventDefault(); button.focus() }
      }
    }
    button.addEventListener('click', close)
    overlay.addEventListener('click', function (event) { if (event.target === overlay) close() })
    window.addEventListener('message', onMessage)
    document.addEventListener('keydown', onKeydown)
    var previousFocus = document.activeElement
    var previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    document.body.appendChild(overlay)
    active = { overlay: overlay, onMessage: onMessage, onKeydown: onKeydown,
      previousFocus: previousFocus, previousOverflow: previousOverflow, onClose: options.onClose }
    button.focus()
    return { close: close }
  }

  function openEmbed(options) {
    var planId = options && String(options.planId || '')
    var publicPlan = /^pl_plan_[0-9a-f]{24}$/.test(planId)
    var legacyPlan = /^[1-9][0-9]*$/.test(planId) && Number.isSafeInteger(Number(planId))
    if (!options || !/^pl_site_[A-Za-z0-9_-]+$/.test(options.siteId || '') || (!publicPlan && !legacyPlan)) {
      throw new Error('Valid siteId and planId are required')
    }
    var api = new URL(options.apiUrl || checkoutOrigin)
    if (api.protocol !== 'https:' && !(api.protocol === 'http:' && (api.hostname === 'localhost' || api.hostname === '127.0.0.1'))) {
      throw new Error('PayLanggan API must use HTTPS')
    }
    var storageKey = 'paylanggan:embed:' + options.siteId + ':' + planId
    function checkoutToken(value) {
      var url = new URL(value)
      if (url.origin !== checkoutOrigin || !/^\/checkout\/[A-Za-z0-9_-]+$/.test(url.pathname)) throw new Error('Invalid PayLanggan checkout URL')
      return url.pathname.split('/')[2]
    }
    function displayOrder(order) {
      checkoutToken(order.checkout_url)
      try { sessionStorage.setItem(storageKey, JSON.stringify(order)) } catch (_) { /* Storage may be unavailable. */ }
      return open({
        checkoutUrl: order.checkout_url,
        onPaid: function () {
          try { sessionStorage.removeItem(storageKey) } catch (_) {}
          document.dispatchEvent(new CustomEvent('paylanggan:paid', { detail: { externalOrderId: order.external_order_id } }))
          if (typeof options.onPaid === 'function') options.onPaid(order.external_order_id)
        },
        onClose: options.onClose,
      })
    }
    function showForm() {
      close()
      var overlay = document.createElement('div')
      overlay.className = 'pl-dropin-overlay'
      overlay.setAttribute('role', 'dialog')
      overlay.setAttribute('aria-modal', 'true')
      overlay.setAttribute('aria-label', 'Mulai pembayaran PayLanggan')
      var panel = document.createElement('div')
      panel.className = 'pl-embed-form-panel'
      panel.innerHTML = '<header class="pl-embed-header"><span class="pl-embed-brand"><img class="pl-embed-logo" src="' + checkoutOrigin + '/paylanggan-logo.png" alt="">PayLanggan</span>' +
        '<button class="pl-embed-close" type="button" aria-label="Tutup formulir"><svg viewBox="0 0 20 20" aria-hidden="true"><path d="m5 5 10 10M15 5 5 15"/></svg></button></header>' +
        '<div class="pl-embed-intro"><span class="pl-embed-eyebrow">MULAI BERLANGGANAN</span>' +
        '<h2>Lanjutkan langganan</h2><p>Lengkapi informasi kontak untuk melanjutkan ke pembayaran.</p></div>' +
        '<form><label>Nama lengkap<input name="name" autocomplete="name" maxlength="100" placeholder="Nama Anda" required></label>' +
        '<label>Alamat email<input name="email" type="email" autocomplete="email" maxlength="100" placeholder="nama@email.com" required></label>' +
        '<label>Nomor ponsel<input name="mobile" type="tel" autocomplete="tel" placeholder="+628123456789" maxlength="20" required></label>' +
        '<p class="pl-embed-error" role="alert"></p><button class="pl-embed-submit" type="submit">Lanjutkan ke pembayaran</button>' +
        '<button class="pl-embed-cancel" type="button">Batal</button></form>'
      var form = panel.querySelector('form')
      var error = panel.querySelector('.pl-embed-error')
      var submit = panel.querySelector('[type="submit"]')
      panel.querySelector('.pl-embed-cancel').addEventListener('click', close)
      panel.querySelector('.pl-embed-close').addEventListener('click', close)
      overlay.appendChild(panel)
      overlay.addEventListener('click', function (event) { if (event.target === overlay) close() })
      var previousFocus = document.activeElement
      var previousOverflow = document.body.style.overflow
      function onKeydown(event) { if (event.key === 'Escape') close() }
      function onMessage() {}
      window.addEventListener('message', onMessage)
      document.addEventListener('keydown', onKeydown)
      document.body.style.overflow = 'hidden'
      document.body.appendChild(overlay)
      // Switching from the contact form to checkout calls close(); it is not a user cancellation.
      active = { overlay: overlay, onMessage: onMessage, onKeydown: onKeydown, previousFocus: previousFocus, previousOverflow: previousOverflow }
      form.querySelector('input').focus()
      form.addEventListener('submit', function (event) {
        event.preventDefault()
        if (submit.disabled) return
        submit.disabled = true
        submit.textContent = 'Menyiapkan pembayaran...'
        error.textContent = ''
        fetch(new URL('/api/v1/embed/sites/' + encodeURIComponent(options.siteId) + '/orders', api), {
          method: 'POST', credentials: 'omit', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(Object.assign(publicPlan ? { plan_public_id: planId } : { plan_id: Number(planId) },
            { customer: Object.fromEntries(new FormData(form)) })),
        }).then(function (response) {
          return response.json().then(function (body) {
            if (!response.ok || !body.success || !body.data || !body.data.checkout_url) throw new Error(body.message || 'Pesanan belum dapat dibuat.')
            return body.data
          })
        }).then(displayOrder).catch(function (reason) {
          error.textContent = reason.message || 'Pembayaran belum tersedia. Coba lagi.'
          submit.disabled = false
          submit.textContent = 'Lanjutkan ke pembayaran'
        })
      })
      return { close: close }
    }
    var saved = null
    try { saved = JSON.parse(sessionStorage.getItem(storageKey) || 'null') } catch (_) {}
    if (!saved || !saved.checkout_url) return Promise.resolve(showForm())
    var token
    try { token = checkoutToken(saved.checkout_url) }
    catch (_) { try { sessionStorage.removeItem(storageKey) } catch (_) {} return Promise.resolve(showForm()) }
    return fetch(new URL('/api/v1/checkout/' + encodeURIComponent(token), api), { credentials: 'omit', cache: 'no-store' })
      .then(function (response) { if (!response.ok) throw new Error('Status pesanan belum dapat diperiksa.'); return response.json() })
      .then(function (body) {
        if (body.data && body.data.invoice_status === 'pending' && ['created', 'active'].indexOf(body.data.checkout_status) !== -1) return displayOrder(saved)
        try { sessionStorage.removeItem(storageKey) } catch (_) {}
        return showForm()
      })
  }

  var style = document.createElement('style')
  style.textContent = '.pl-dropin-overlay{position:fixed;inset:0;z-index:2147483647;display:flex;align-items:center;justify-content:center;padding:16px;background:rgba(15,23,42,.7)}' +
    '.pl-dropin-panel{position:relative;width:min(100%,1050px);height:min(92vh,850px);overflow:hidden;border-radius:16px;background:#fff;box-shadow:0 24px 80px rgba(0,0,0,.28)}' +
    '.pl-dropin-panel iframe{display:block;width:100%;height:100%;border:0}' +
    '.pl-dropin-controls{position:absolute;top:10px;right:10px;z-index:2;display:flex;align-items:center;gap:8px;padding:4px;border-radius:24px;background:rgba(255,255,255,.97);box-shadow:0 2px 12px rgba(15,23,42,.16)}' +
    '.pl-dropin-close{position:static;display:grid;place-items:center;flex:0 0 36px;width:36px;height:36px;padding:0;border:0;border-radius:50%;background:#f1f5f9;color:#111827;font:26px/1 system-ui,sans-serif;cursor:pointer}' +
    '.pl-dropin-close:hover{background:#e2e8f0}.pl-dropin-close:focus-visible,.pl-dropin-full-page:focus-visible{outline:2px solid #3155f5;outline-offset:2px}' +
    '.pl-dropin-full-page{display:inline-flex;align-items:center;justify-content:center;min-height:36px;padding:0 12px;border-radius:18px;background:#fff;color:#2455eb;font:600 12px system-ui,sans-serif;white-space:nowrap;text-decoration:none}' +
    '@media(max-width:640px){.pl-dropin-overlay{padding:0}.pl-dropin-panel{height:100vh;width:100%;border-radius:0}}' +
    '.pl-embed-form-panel{box-sizing:border-box;width:min(100%,460px);max-height:90vh;overflow:auto;border:1px solid rgba(148,163,184,.2);border-radius:22px;background:#fff;padding:24px 28px 26px;color:#172033;font:14px/1.5 system-ui,-apple-system,"Segoe UI",sans-serif;box-shadow:0 28px 90px rgba(15,23,42,.3)}' +
    '.pl-embed-form-panel *{box-sizing:border-box}.pl-embed-header{display:flex;align-items:center;justify-content:space-between;padding-bottom:20px;border-bottom:1px solid #eef1f6}' +
    '.pl-embed-brand{display:inline-flex;align-items:center;gap:9px;font-size:15px;font-weight:750;letter-spacing:-.03em;color:#111827}.pl-embed-logo{display:block;width:30px;height:30px;object-fit:contain}' +
    '.pl-embed-form-panel button.pl-embed-close{appearance:none;box-sizing:border-box;position:relative;display:flex;flex:0 0 36px;align-items:center;justify-content:center;width:36px!important;min-width:36px;height:36px;min-height:36px;margin:0!important;padding:0!important;overflow:hidden;border:0;border-radius:50%;background:#f1f5f9;color:#475569;line-height:1;cursor:pointer}' +
    '.pl-embed-form-panel button.pl-embed-close:hover{background:#e2e8f0;color:#111827}.pl-embed-close svg{display:block;flex:none;width:16px;height:16px;overflow:visible}.pl-embed-close path{fill:none;stroke:currentColor;stroke-width:1.8;stroke-linecap:round}' +
    '.pl-embed-intro{padding:23px 0 4px}.pl-embed-eyebrow{color:#3155f5;font-size:10px;font-weight:750;letter-spacing:.12em}.pl-embed-form-panel h2{margin:7px 0 4px;font-size:25px;line-height:1.2;letter-spacing:-.04em}.pl-embed-form-panel .pl-embed-intro p{margin:0;color:#64748b;font-size:13px}' +
    '.pl-embed-form-panel label{display:block;margin:16px 0 0;color:#263247;font-size:12px;font-weight:650}.pl-embed-form-panel input{display:block;width:100%;height:46px;margin-top:7px;border:1px solid #d7deea;border-radius:11px;background:#fff;padding:0 13px;color:#111827;font:14px system-ui,sans-serif;outline:none;transition:border-color .15s,box-shadow .15s}.pl-embed-form-panel input::placeholder{color:#9aa5b5}.pl-embed-form-panel input:focus{border-color:#4564ff;box-shadow:0 0 0 3px rgba(49,85,245,.13)}' +
    '.pl-embed-form-panel button:focus-visible{outline:3px solid rgba(49,85,245,.28);outline-offset:2px}.pl-embed-form-panel .pl-embed-submit,.pl-embed-form-panel .pl-embed-cancel{display:block;width:100%;min-height:45px;margin-top:18px;padding:11px 14px;border:0;border-radius:11px;font:650 13px system-ui,sans-serif;cursor:pointer;transition:background .15s,transform .15s}.pl-embed-form-panel .pl-embed-submit{background:#3155f5;color:#fff;box-shadow:0 8px 18px rgba(49,85,245,.2)}.pl-embed-form-panel .pl-embed-submit:hover{background:#2447e8}.pl-embed-form-panel .pl-embed-cancel{min-height:40px;margin-top:8px;background:transparent;color:#64748b}.pl-embed-form-panel .pl-embed-cancel:hover{background:#f4f6fa;color:#334155}.pl-embed-form-panel button:disabled{opacity:.6;cursor:wait}.pl-embed-form-panel .pl-embed-error{margin:10px 0 0;color:#b42318;font-size:12px;min-height:0}.pl-embed-form-panel .pl-embed-error:empty{display:none}'
  document.head.appendChild(style)
  window.PayLanggan = Object.freeze({ open: open, openEmbed: openEmbed, close: close })
  var siteId = script.getAttribute('data-pl-site')
  var apiUrl = script.getAttribute('data-pl-api-url')
  if (siteId && apiUrl) {
    document.addEventListener('click', function (event) {
      var button = event.target && event.target.closest && event.target.closest('[data-pl-checkout]')
      if (!button) return
      event.preventDefault()
      openEmbed({ siteId: siteId, planId: button.getAttribute('data-pl-plan'), apiUrl: apiUrl }).catch(function (reason) {
        document.dispatchEvent(new CustomEvent('paylanggan:error', { detail: { message: reason.message || 'Pembayaran belum tersedia.' } }))
      })
    })
  }
})()
