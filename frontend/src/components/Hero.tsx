import { Button } from "@/components/ui/button";
import { ArrowRight, Play } from "lucide-react";
import heroIllustration from "@/assets/hero-illustration.png";

const Hero = () => {
  return (
    <section className="pt-24 pb-16 lg:pt-32 lg:pb-24 overflow-hidden">
      <div className="container mx-auto px-4">
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          <div className="space-y-8">
            <div className="inline-flex items-center gap-2 bg-accent px-4 py-2 rounded-full">
              <span className="w-2 h-2 bg-accent-foreground rounded-full animate-pulse"></span>
              <span className="text-sm text-accent-foreground font-medium">Platform katalog produk #1 di Indonesia</span>
            </div>
            
            <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-foreground leading-tight">
              Buat Katalog Produk, 
              <span className="text-primary"> Langsung Jualan</span> via WhatsApp
            </h1>
            
            <p className="text-lg text-muted-foreground max-w-xl">
              Sello membantu Anda membuat katalog produk online yang menarik dan mengarahkan calon pembeli langsung ke WhatsApp atau link tujuan Anda. Lengkap dengan analitik klik untuk memahami performa produk.
            </p>
            
            <div className="flex flex-col sm:flex-row gap-4">
              <Button variant="hero" size="xl">
                Mulai Gratis Sekarang
                <ArrowRight className="ml-2" size={20} />
              </Button>
              <Button variant="heroOutline" size="xl">
                <Play className="mr-2" size={20} />
                Lihat Demo
              </Button>
            </div>
            
            <div className="flex items-center gap-8 pt-4">
              <div>
                <p className="text-3xl font-bold text-foreground">10K+</p>
                <p className="text-sm text-muted-foreground">Toko Aktif</p>
              </div>
              <div className="w-px h-12 bg-border"></div>
              <div>
                <p className="text-3xl font-bold text-foreground">1M+</p>
                <p className="text-sm text-muted-foreground">Produk Dikatalog</p>
              </div>
              <div className="w-px h-12 bg-border"></div>
              <div>
                <p className="text-3xl font-bold text-foreground">99%</p>
                <p className="text-sm text-muted-foreground">Uptime</p>
              </div>
            </div>
          </div>
          
          <div className="relative">
            <div className="absolute -inset-4 bg-gradient-to-r from-primary/20 to-accent/40 rounded-3xl blur-3xl opacity-50"></div>
            <img 
              src={heroIllustration} 
              alt="Sello Dashboard Preview" 
              className="relative rounded-2xl shadow-2xl w-full"
            />
          </div>
        </div>
      </div>
    </section>
  );
};

export default Hero;
