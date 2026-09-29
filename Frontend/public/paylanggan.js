/* PayLanggan drop-in checkout. Load from the same origin as the checkout page. */
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
    panel.appendChild(fullPage)
    panel.appendChild(button)
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

  var style = document.createElement('style')
  style.textContent = '.pl-dropin-overlay{position:fixed;inset:0;z-index:2147483647;display:flex;align-items:center;justify-content:center;padding:16px;background:rgba(15,23,42,.7)}' +
    '.pl-dropin-panel{position:relative;width:min(100%,1050px);height:min(92vh,850px);overflow:hidden;border-radius:16px;background:#fff;box-shadow:0 24px 80px rgba(0,0,0,.28)}' +
    '.pl-dropin-panel iframe{display:block;width:100%;height:100%;border:0}' +
    '.pl-dropin-close{position:absolute;right:10px;top:8px;z-index:2;width:36px;height:36px;border:0;border-radius:50%;background:#fff;color:#111827;font-size:27px;line-height:1;cursor:pointer;box-shadow:0 1px 8px rgba(0,0,0,.16)}' +
    '.pl-dropin-full-page{position:absolute;right:54px;top:13px;z-index:2;padding:3px 8px;border-radius:5px;background:#fff;color:#2455eb;font:12px system-ui,sans-serif;text-decoration:underline}' +
    '@media(max-width:640px){.pl-dropin-overlay{padding:0}.pl-dropin-panel{height:100vh;width:100%;border-radius:0}}'
  document.head.appendChild(style)
  window.PayLanggan = Object.freeze({ open: open, close: close })
})()
