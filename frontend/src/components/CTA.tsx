import { Button } from "@/components/ui/button";
import { ArrowRight } from "lucide-react";

const CTA = () => {
  return (
    <section className="py-20">
      <div className="container mx-auto px-4">
        <div className="relative overflow-hidden bg-gradient-to-r from-primary via-[hsl(170,70%,50%)] to-[hsl(145,85%,55%)] rounded-3xl p-12 md:p-16 text-center">
          {/* Background decoration */}
          <div className="absolute inset-0 opacity-10">
            <div className="absolute top-0 left-0 w-72 h-72 bg-primary-foreground rounded-full blur-3xl -translate-x-1/2 -translate-y-1/2"></div>
            <div className="absolute bottom-0 right-0 w-96 h-96 bg-primary-foreground rounded-full blur-3xl translate-x-1/2 translate-y-1/2"></div>
          </div>
          
          <div className="relative z-10 max-w-2xl mx-auto">
            <h2 className="text-3xl md:text-4xl lg:text-5xl font-bold text-primary-foreground mb-6">
              Siap Tingkatkan Penjualan Anda?
            </h2>
            <p className="text-lg text-primary-foreground/80 mb-8">
              Bergabung dengan ribuan penjual yang sudah menggunakan Sello untuk meningkatkan konversi dan penjualan mereka.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button 
                size="xl" 
                className="bg-primary-foreground text-primary hover:bg-primary-foreground/90"
              >
                Mulai Gratis Sekarang
                <ArrowRight className="ml-2" size={20} />
              </Button>
              <Button 
                size="xl" 
                variant="ghost"
                className="text-primary-foreground border-2 border-primary-foreground/30 hover:bg-primary-foreground/10"
              >
                Jadwalkan Demo
              </Button>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
};

export default CTA;
