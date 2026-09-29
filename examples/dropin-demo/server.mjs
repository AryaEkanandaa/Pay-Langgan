import { createServer } from 'node:http'
import { randomUUID } from 'node:crypto'
import { readFile } from 'node:fs/promises'

const apiUrl = process.env.PAYLANGGAN_API_URL || 'http://localhost:8080'
const checkoutOrigin = process.env.PAYLANGGAN_CHECKOUT_ORIGIN || 'http://localhost:5173'
const secretKey = process.env.PAYLANGGAN_SECRET_KEY
const planId = Number(process.env.PAYLANGGAN_PLAN_ID)
const port = Number(process.env.PORT || 4174)

if (!secretKey || !Number.isInteger(planId) || planId < 1) {
  throw new Error('Set PAYLANGGAN_SECRET_KEY and PAYLANGGAN_PLAN_ID before starting the demo.')
}

async function readJSON(request) {
  let body = ''
  for await (const chunk of request) {
    body += chunk
    if (body.length > 8192) throw new Error('Request too large')
  }
  return JSON.parse(body)
}

function json(response, status, data) {
  response.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8', 'Cache-Control': 'no-store' })
  response.end(JSON.stringify(data))
}

async function paylanggan(path, options = {}) {
  const upstream = await fetch(new URL(path, apiUrl), {
    ...options,
    headers: { Authorization: `Bearer ${secretKey}`, 'Content-Type': 'application/json' },
  })
  const data = await upstream.json()
  if (!upstream.ok) throw new Error(data.message || `PayLanggan HTTP ${upstream.status}`)
  return data.data
}

createServer(async (request, response) => {
  try {
    const url = new URL(request.url, `http://localhost:${port}`)
    if (request.method === 'GET' && url.pathname === '/') {
      const html = await readFile(new URL('./index.html', import.meta.url), 'utf8')
      response.writeHead(200, { 'Content-Type': 'text/html; charset=utf-8', 'Cache-Control': 'no-store' })
      response.end(html.replaceAll('__WIDGET_URL__', new URL('/paylanggan.js', checkoutOrigin).toString()))
      return
    }
    if (request.method === 'POST' && url.pathname === '/api/order') {
      const customer = await readJSON(request)
      const externalOrderId = randomUUID()
      const order = await paylanggan('/api/v1/integrations/orders', {
        method: 'POST',
        body: JSON.stringify({ external_order_id: externalOrderId, plan_id: planId, customer }),
      })
      json(response, 201, order)
      return
    }
    if (request.method === 'GET' && url.pathname.startsWith('/api/order/')) {
      const reference = encodeURIComponent(url.pathname.slice('/api/order/'.length))
      json(response, 200, await paylanggan(`/api/v1/integrations/orders/${reference}`))
      return
    }
    json(response, 404, { error: 'Not found' })
  } catch (error) {
    json(response, 400, { error: error instanceof Error ? error.message : 'Request failed' })
  }
}).listen(port, () => console.log(`Merchant demo: http://localhost:${port}`))
