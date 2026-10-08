const steps = [
  {
    number: "01",
    title: "Daftar & Buat Toko",
    description: "Daftar gratis dalam hitungan detik dan buat toko online Anda dengan subdomain sello.id"
  },
  {
    number: "02", 
    title: "Tambah Produk",
    description: "Upload foto, tulis deskripsi, dan atur harga produk Anda dengan mudah"
  },
  {
    number: "03",
    title: "Atur Link Tujuan",
    description: "Pilih kemana pembeli akan diarahkan - WhatsApp, marketplace, atau link custom"
  },
  {
    number: "04",
    title: "Bagikan & Pantau",
    description: "Bagikan link katalog Anda dan pantau performa dengan analitik lengkap"
  }
];

const HowItWorks = () => {
  return (
    <section id="cara-kerja" className="py-20">
      <div className="container mx-auto px-4">
        <div className="text-center max-w-2xl mx-auto mb-16">
          <span className="text-primary font-medium">Cara Kerja</span>
          <h2 className="text-3xl md:text-4xl font-bold text-foreground mt-2 mb-4">
            Mulai Jualan dalam 4 Langkah Mudah
          </h2>
          <p className="text-muted-foreground">
            Tidak perlu keahlian teknis. Buat katalog produk profesional dalam hitungan menit.
          </p>
        </div>
        
        <div className="relative">
          {/* Connection Line */}
          <div className="hidden lg:block absolute top-1/2 left-0 right-0 h-0.5 bg-border -translate-y-1/2"></div>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            {steps.map((step, index) => (
              <div key={index} className="relative">
              <div className="bg-card p-8 rounded-2xl border border-border text-center relative z-10">
                <div className="w-16 h-16 bg-gradient-to-r from-primary to-[hsl(145,85%,55%)] rounded-full flex items-center justify-center mx-auto mb-6 shadow-lg">
                  <span className="text-2xl font-bold text-primary-foreground">{step.number}</span>
                  </div>
                  <h3 className="text-xl font-semibold text-foreground mb-3">
                    {step.title}
                  </h3>
                  <p className="text-muted-foreground">
                    {step.description}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};

export default HowItWorks;
