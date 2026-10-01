import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import vm from 'node:vm'

const listeners = new Map()
function element(tag) {
  const item = {
    tagName: tag, style: {}, children: [], attributes: {}, contentWindow: tag === 'iframe' ? {} : undefined,
    appendChild(child) { this.children.push(child) },
    remove() { this.removed = true },
    setAttribute(name, value) { this.attributes[name] = value },
    addEventListener(name, handler) { this[`on${name}`] = handler },
    focus() {},
  }
  return item
}
const body = element('body')
const document = {
  currentScript: { src: 'https://checkout.example/paylanggan.js', getAttribute() { return null } },
  head: element('head'), body, activeElement: element('button'),
  createElement: element,
  addEventListener(name, handler) { listeners.set(`document:${name}`, handler) },
  removeEventListener(name) { listeners.delete(`document:${name}`) },
}
const window = {
  location: { origin: 'https://merchant.example' },
  addEventListener(name, handler) { listeners.set(name, handler) },
  removeEventListener(name) { listeners.delete(name) },
}
vm.runInNewContext(await readFile(new URL('../../Frontend/public/paylanggan.js', import.meta.url), 'utf8'),
  { window, document, URL })

assert.ok(window.PayLanggan)
assert.throws(() => window.PayLanggan.open({ checkoutUrl: 'https://other.example/checkout/token' }), /Invalid/)
let paid = 0
const modal = window.PayLanggan.open({
  checkoutUrl: 'https://checkout.example/checkout/abcDEF_123',
  onPaid() { paid++ },
})
const overlay = body.children.at(-1)
const frame = overlay.children[0].children.find((child) => child.tagName === 'iframe')
assert.equal(new URL(frame.src).searchParams.get('parent_origin'), 'https://merchant.example')
listeners.get('message')({ origin: 'https://other.example', source: frame.contentWindow,
  data: { type: 'paylanggan:checkout-status', status: 'paid' } })
assert.equal(paid, 0)
listeners.get('message')({ origin: 'https://checkout.example', source: frame.contentWindow,
  data: { type: 'paylanggan:checkout-status', status: 'paid' } })
assert.equal(paid, 1)
modal.close()
assert.equal(overlay.removed, true)
assert.equal(listeners.has('message'), false)
console.log('Drop-in widget smoke test passed')
