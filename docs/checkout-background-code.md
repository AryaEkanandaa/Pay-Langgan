# Latar checkout terpadu

Satu gradient dekoratif berada di checkout-shell, container terluar React yang membungkus header, stepper, form, dan footer dalam iframe modal. Wrapper stepper/form/footer transparan. Header tetap putih pada mode terang dan gelap pada dark mode, dengan pemisah lembut. Gradient tidak ikut scroll karena yang digulir hanya area form. Tidak ada dependency atau perubahan alur pembayaran.

## CSS yang berubah (Frontend/src/index.css)

```css
.checkout-shell {
  --modal-bg: #eef1fb;
  --modal-glow-1: rgba(54, 84, 255, 0.35);
  --modal-glow-2: rgba(255, 183, 3, 0.3);
  --modal-glow-3: rgba(155, 92, 255, 0.28);
  background-color: var(--modal-bg);
  background-image:
    radial-gradient(38% 32% at 12% 8%, var(--modal-glow-1), transparent 70%),
    radial-gradient(34% 30% at 88% 12%, var(--modal-glow-2), transparent 70%),
    radial-gradient(42% 38% at 50% 100%, var(--modal-glow-3), transparent 70%);
  background-position: center;
  background-repeat: no-repeat;
  background-size: 100% 100%;
}
html.dark .checkout-shell {
  --modal-bg: #13192a;
  --modal-glow-1: rgba(54, 84, 255, 0.18);
  --modal-glow-2: rgba(255, 183, 3, 0.08);
  --modal-glow-3: rgba(155, 92, 255, 0.14);
}
.checkout-stepper-header {
  background: transparent;
  backdrop-filter: none;
  box-shadow: none;
  border-bottom: 0;
}
.checkout-scroll, .checkout-footer { background: transparent; }
.checkout-footer { padding-bottom: env(safe-area-inset-bottom, 0px); }
.checkout-scroll-fade { background: linear-gradient(transparent, var(--modal-bg)); }

```

## Wrapper yang berubah (Frontend/src/pages/CheckoutPage.tsx)

```tsx
<main className="checkout-shell flex h-dvh min-h-0 flex-col overflow-hidden text-ink">
  {/* Header, stepper dan isi checkout tetap menggunakan komponen yang sama. */}
  {/* Footer tidak lagi memakai border atau warna latar terpisah. */}
  <footer className="checkout-footer z-10 shrink-0">
    {/* Aksi pembayaran dan informasi keamanan yang sudah ada. */}
  </footer>
</main>
```
