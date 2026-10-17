import { createServer, type IncomingMessage, type ServerResponse } from 'node:http'
import { renderEmail } from '../src/emails/render'
import type { EmailType, InvoiceIssuedProps, PaymentFailedProps, PaymentSuccessProps, TrialStartedProps } from '../src/emails/types'

const maxRequestBytes = 256 * 1024
const port = Number(process.env.EMAIL_RENDER_PORT ?? '3101')
if (!Number.isInteger(port) || port < 1 || port > 65535) throw new Error('EMAIL_RENDER_PORT tidak valid')

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function validBase(value: Record<string, unknown>): boolean {
  return isObject(value.brand) && typeof value.brand.name === 'string'
    && isObject(value.customer) && typeof value.customer.name === 'string' && typeof value.customer.email === 'string'
}

function validInvoiceProps(value: unknown): value is InvoiceIssuedProps {
  if (!isObject(value) || !validBase(value) || !isObject(value.invoice)) return false
  return typeof value.invoice.number === 'string' && typeof value.invoice.total === 'number'
    && Array.isArray(value.invoice.items) && typeof value.invoice.url === 'string'
}

function validTrialProps(value: unknown): value is TrialStartedProps {
  if (!isObject(value) || !validBase(value) || !isObject(value.trial)) return false
  const trial = value.trial
  return typeof trial.plan === 'string' && typeof trial.endsAt === 'string'
    && (trial.billingCycle === 'monthly' || trial.billingCycle === 'yearly')
    && typeof trial.amount === 'number' && typeof trial.autoDebitEnabled === 'boolean'
    && Array.isArray(trial.items) && trial.items.every((item: unknown) => isObject(item)
      && typeof item.name === 'string' && typeof item.quantity === 'number' && typeof item.subtotal === 'number')
}

async function readJson(request: IncomingMessage): Promise<unknown> {
  const chunks: Buffer[] = []
  let size = 0
  for await (const chunk of request) {
    const bytes = Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk)
    size += bytes.length
    if (size > maxRequestBytes) throw new Error('Permintaan terlalu besar')
    chunks.push(bytes)
  }
  return JSON.parse(Buffer.concat(chunks).toString('utf8')) as unknown
}

function respond(response: ServerResponse, status: number, body: object): void {
  response.writeHead(status, { 'Content-Type': 'application/json; charset=utf-8', 'Cache-Control': 'no-store', 'X-Content-Type-Options': 'nosniff' })
  response.end(JSON.stringify(body))
}

const server = createServer(async (request, response) => {
  if (request.method === 'GET' && request.url === '/health') {
    respond(response, 200, { status: 'ok' })
    return
  }
  if (request.method !== 'POST' || request.url !== '/render' || !request.headers['content-type']?.startsWith('application/json')) {
    respond(response, 404, { error: 'Not found' })
    return
  }
  try {
    const payload = await readJson(request)
    if (!isObject(payload) || !isObject(payload.props)) {
      respond(response, 422, { error: 'Data email tidak valid' })
      return
    }
    const type = payload.type as EmailType
    let result
    switch (type) {
      case 'invoice-issued':
        if (!validInvoiceProps(payload.props)) {
          respond(response, 422, { error: 'Data email tidak valid' })
          return
        }
        result = await renderEmail(type, payload.props)
        break
      case 'payment-success':
        if (!validInvoiceProps(payload.props)) {
          respond(response, 422, { error: 'Data email tidak valid' })
          return
        }
        result = await renderEmail(type, payload.props as PaymentSuccessProps)
        break
      case 'payment-failed':
        if (!validInvoiceProps(payload.props)) {
          respond(response, 422, { error: 'Data email tidak valid' })
          return
        }
        if (typeof (payload.props as PaymentFailedProps).failureReason !== 'string' || typeof (payload.props as PaymentFailedProps).actionUrl !== 'string') {
          respond(response, 422, { error: 'Data email tidak valid' })
          return
        }
        result = await renderEmail(type, payload.props as PaymentFailedProps)
        break
      case 'trial-started':
        if (!validTrialProps(payload.props)) {
          respond(response, 422, { error: 'Data email tidak valid' })
          return
        }
        result = await renderEmail(type, payload.props)
        break
      default:
        respond(response, 422, { error: 'Jenis email tidak didukung' })
        return
    }
    respond(response, 200, result)
  } catch (error) {
    process.stderr.write(`Email render gagal: ${error instanceof Error ? error.message : 'unknown error'}\n`)
    respond(response, 422, { error: 'Email belum dapat dirender' })
  }
})

server.requestTimeout = 10_000
server.listen(port, '127.0.0.1', () => {
  process.stdout.write(`Email renderer siap di http://127.0.0.1:${port}\n`)
})
