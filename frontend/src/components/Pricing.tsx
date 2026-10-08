import { Button } from "@/components/ui/button";
import { Check } from "lucide-react";

const plans = [
  {
    name: "Gratis",
    price: "Rp 0",
    period: "selamanya",
    description: "Cocok untuk memulai bisnis online Anda",
    features: [
      "Subdomain gratis (toko.sello.id)",
      "Hingga 20 produk",
      "Analitik dasar",
      "Integrasi WhatsApp",
      "Template standar"
    ],
    cta: "Mulai Gratis",
    popular: false
  },
  {
    name: "Pro",
    price: "Rp 49K",
    period: "/bulan",
    description: "Untuk bisnis yang ingin berkembang",
    features: [
      "Custom domain Anda sendiri",
      "Produk unlimited",
      "Analitik lengkap & export",
      "Multi link tujuan",
      "Semua template premium",
      "Hapus branding Sello",
      "Prioritas support"
    ],
    cta: "Pilih Pro",
    popular: true
  },
  {
    name: "Business",
    price: "Rp 149K",
    period: "/bulan",
    description: "Untuk tim dan bisnis besar",
    features: [
      "Semua fitur Pro",
      "5 user/admin",
      "API access",
      "White label",
      "Custom branding",
      "Dedicated support",
      "SLA 99.9%"
    ],
    cta: "Hubungi Sales",
    popular: false
  }
];

const Pricing = () => {
  return (
    <section id="harga" className="py-20 bg-card">
      <div className="container mx-auto px-4">
        <div className="text-center max-w-2xl mx-auto mb-16">
          <span className="text-primary font-medium">Harga</span>
          <h2 className="text-3xl md:text-4xl font-bold text-foreground mt-2 mb-4">
            Pilih Paket yang Sesuai Kebutuhan Anda
          </h2>
          <p className="text-muted-foreground">
            Mulai gratis, upgrade kapan saja. Tanpa biaya tersembunyi.
          </p>
        </div>
        
        <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
          {plans.map((plan, index) => (
            <div 
              key={index}
              className={`relative bg-background p-8 rounded-2xl border-2 transition-all duration-300 hover:shadow-xl ${
                plan.popular 
                  ? 'border-primary shadow-lg scale-105' 
                  : 'border-border hover:border-primary/50'
              }`}
            >
              {plan.popular && (
                <div className="absolute -top-4 left-1/2 -translate-x-1/2 bg-primary text-primary-foreground px-4 py-1 rounded-full text-sm font-medium">
                  Paling Populer
                </div>
              )}
              
              <div className="text-center mb-8">
                <h3 className="text-xl font-semibold text-foreground mb-2">{plan.name}</h3>
                <div className="flex items-baseline justify-center gap-1">
                  <span className="text-4xl font-bold text-foreground">{plan.price}</span>
                  <span className="text-muted-foreground">{plan.period}</span>
                </div>
                <p className="text-muted-foreground mt-2">{plan.description}</p>
              </div>
              
              <ul className="space-y-4 mb-8">
                {plan.features.map((feature, featureIndex) => (
                  <li key={featureIndex} className="flex items-start gap-3">
                    <div className="w-5 h-5 rounded-full bg-accent flex items-center justify-center flex-shrink-0 mt-0.5">
                      <Check className="w-3 h-3 text-accent-foreground" />
                    </div>
                    <span className="text-foreground">{feature}</span>
                  </li>
                ))}
              </ul>
              
              <Button 
                className="w-full" 
                variant={plan.popular ? "default" : "outline"}
                size="lg"
              >
                {plan.cta}
              </Button>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default Pricing;
