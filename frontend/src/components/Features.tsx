import { 
  MessageCircle, 
  BarChart3, 
  Globe, 
  Palette, 
  Link2, 
  Zap 
} from "lucide-react";

const features = [
  {
    icon: MessageCircle,
    title: "Integrasi WhatsApp",
    description: "Arahkan pembeli langsung ke WhatsApp dengan pesan otomatis yang sudah terisi nama produk dan harga."
  },
  {
    icon: BarChart3,
    title: "Analitik Lengkap",
    description: "Pantau jumlah klik produk, CTA, dan performa katalog Anda dengan dashboard analitik real-time."
  },
  {
    icon: Globe,
    title: "Subdomain Gratis",
    description: "Dapatkan subdomain gratis seperti toko-anda.sello.id atau upgrade ke domain custom Anda sendiri."
  },
  {
    icon: Palette,
    title: "Desain Kustomisasi",
    description: "Pilih dari berbagai tema menarik atau sesuaikan tampilan katalog sesuai brand Anda."
  },
  {
    icon: Link2,
    title: "Multi Link Tujuan",
    description: "Arahkan ke WhatsApp, Shopee, Tokopedia, atau link apapun yang Anda inginkan."
  },
  {
    icon: Zap,
    title: "Super Cepat",
    description: "Katalog Anda dimuat dengan cepat di semua perangkat untuk pengalaman pembeli yang optimal."
  }
];

const Features = () => {
  return (
    <section id="fitur" className="py-20 bg-card">
      <div className="container mx-auto px-4">
        <div className="text-center max-w-2xl mx-auto mb-16">
          <span className="text-primary font-medium">Fitur Unggulan</span>
          <h2 className="text-3xl md:text-4xl font-bold text-foreground mt-2 mb-4">
            Semua yang Anda Butuhkan untuk Berjualan Online
          </h2>
          <p className="text-muted-foreground">
            Sello menyediakan fitur lengkap untuk membuat katalog produk profesional yang mendorong konversi.
          </p>
        </div>
        
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <div 
              key={index}
              className="bg-background p-8 rounded-2xl border border-border hover:border-primary/50 hover:shadow-xl transition-all duration-300 group"
            >
              <div className="w-14 h-14 bg-accent rounded-xl flex items-center justify-center mb-6 group-hover:bg-gradient-to-r group-hover:from-primary group-hover:to-[hsl(145,85%,55%)] group-hover:scale-110 transition-all duration-300">
                <feature.icon className="w-7 h-7 text-accent-foreground group-hover:text-primary-foreground transition-colors" />
              </div>
              <h3 className="text-xl font-semibold text-foreground mb-3">
                {feature.title}
              </h3>
              <p className="text-muted-foreground">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default Features;
