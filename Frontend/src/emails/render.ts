import { createElement } from 'react'
import { render, toPlainText } from '@react-email/render'
import { emailRegistry, type EmailRegistryEntry } from './registry'
import type { EmailPropsByType, EmailType } from './types'

export interface RenderedEmail {
  subject: string
  html: string
  text: string
}

export async function renderEmail<K extends EmailType>(type: K, props: EmailPropsByType[K]): Promise<RenderedEmail> {
  const entry = emailRegistry[type] as EmailRegistryEntry<EmailPropsByType[K]>
  const subject = entry.subject(props)
  const html = await render(createElement(entry.component, props))
  return { subject, html, text: toPlainText(html) }
}
