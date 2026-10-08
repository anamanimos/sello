import { Button } from "@/components/ui/button";
import { Menu, X } from "lucide-react";
import { useState } from "react";

const Header = () => {
  const [isMenuOpen, setIsMenuOpen] = useState(false);

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-md border-b border-border">
      <div className="container mx-auto px-4 py-4 flex items-center justify-between">
        <a href="/" className="flex items-center gap-2">
          <div className="w-8 h-8 bg-gradient-to-r from-primary to-[hsl(145,85%,55%)] rounded-lg flex items-center justify-center shadow-md">
            <span className="text-primary-foreground font-bold text-lg">S</span>
          </div>
          <span className="text-xl font-bold text-foreground">Sello</span>
        </a>

        {/* Desktop Navigation */}
        <nav className="hidden md:flex items-center gap-8">
          <a href="#fitur" className="text-muted-foreground hover:text-foreground transition-colors">
            Fitur
          </a>
          <a href="#cara-kerja" className="text-muted-foreground hover:text-foreground transition-colors">
            Cara Kerja
          </a>
          <a href="#harga" className="text-muted-foreground hover:text-foreground transition-colors">
            Harga
          </a>
        </nav>

        <div className="hidden md:flex items-center gap-4">
          <Button variant="ghost">Masuk</Button>
          <Button>Daftar Gratis</Button>
        </div>

        {/* Mobile Menu Button */}
        <button
          className="md:hidden p-2"
          onClick={() => setIsMenuOpen(!isMenuOpen)}
        >
          {isMenuOpen ? <X size={24} /> : <Menu size={24} />}
        </button>
      </div>

      {/* Mobile Navigation */}
      {isMenuOpen && (
        <div className="md:hidden bg-background border-b border-border">
          <nav className="container mx-auto px-4 py-4 flex flex-col gap-4">
            <a href="#fitur" className="text-muted-foreground hover:text-foreground transition-colors">
              Fitur
            </a>
            <a href="#cara-kerja" className="text-muted-foreground hover:text-foreground transition-colors">
              Cara Kerja
            </a>
            <a href="#harga" className="text-muted-foreground hover:text-foreground transition-colors">
              Harga
            </a>
            <div className="flex flex-col gap-2 pt-4 border-t border-border">
              <Button variant="ghost" className="w-full">Masuk</Button>
              <Button className="w-full">Daftar Gratis</Button>
            </div>
          </nav>
        </div>
      )}
    </header>
  );
};

export default Header;
