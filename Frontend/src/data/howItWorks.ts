export interface Step {
  number: string
  title: string
  description: string
}

export const steps: Step[] = [
  {
    number: '01',
    title: 'Susun Katalog',
    description:
      'Buat layanan, produk, plan, dan add-on sesuai model bisnis Anda sendiri.',
  },
  {
    number: '02',
    title: 'Daftarkan Pelanggan',
    description:
      'Tambahkan data pelanggan lalu buat langganan dari plan yang sudah tersedia.',
  },
  {
    number: '03',
    title: 'Kelola Langganan',
    description:
      'Terapkan kupon, tambahkan add-on, atau jeda dan batalkan langganan kapan saja.',
  },
]
