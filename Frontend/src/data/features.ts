export interface Feature {
  title: string
  description: string
  span: string
}

export const features: Feature[] = [
  {
    title: 'Siklus Hidup Langganan',
    description:
      'Kelola langganan pelanggan secara penuh trial, aktif, jeda, lanjutkan, hingga batal lengkap dengan add-on dan kupon per langganan.',
    span: 'md:col-span-8',
  },
  {
    title: 'Identitas & Multi-Tenant',
    description:
      'Registrasi bisnis dan autentikasi berbasis JWT, siap dipakai untuk banyak tenant sekaligus.',
    span: 'md:col-span-4',
  },
  {
    title: 'Katalog Produk & Plan',
    description:
      'Susun layanan, produk, paket harga, dan add-on secara fleksibel sesuai model bisnis masing-masing tenant.',
    span: 'md:col-span-3',
  },
  {
    title: 'Manajemen Pelanggan',
    description:
      'Simpan data pelanggan dan pantau riwayat langganan mereka dari satu daftar terpusat.',
    span: 'md:col-span-3',
  },
  {
    title: 'Kupon & Diskon',
    description:
      'Buat kode kupon dan terapkan langsung ke langganan pelanggan saat pembuatan atau perpanjangan.',
    span: 'md:col-span-3',
  },
  {
    title: 'Fondasi Payment Gateway',
    description:
      'Struktur provider pembayaran sudah disiapkan di level kode, siap dihubungkan ke penyedia penagihan digital.',
    span: 'md:col-span-3',
  },
]
