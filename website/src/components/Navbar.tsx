"use client";

import { useEffect, useState } from "react";
import { Server, Key, Menu, X } from "lucide-react";
import { usePathname } from "next/navigation";
import Link from "next/link";

export default function Navbar() {
  const [activeSection, setActiveSection] = useState("hero");
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const pathname = usePathname();

  // Helper hook to identify current application view
  const isHome = pathname === "/";

  useEffect(() => {
    if (!isHome) return;

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            setActiveSection(entry.target.id);
          }
        });
      },
      { rootMargin: "-20% 0px -60% 0px" }
    );

    // Minor delay to ensure DOM is fully repainted after route transition
    const timeout = setTimeout(() => {
      const sections = document.querySelectorAll("section[id]");
      sections.forEach((s) => observer.observe(s));
    }, 100);
    
    return () => {
      clearTimeout(timeout);
      observer.disconnect();
    };
  }, [isHome]);

  // Lock body scroll when mobile menu is open
  useEffect(() => {
    if (isMobileMenuOpen) {
      document.body.style.overflow = "hidden";
    } else {
      document.body.style.overflow = "";
    }
    return () => {
      document.body.style.overflow = "";
    };
  }, [isMobileMenuOpen]);

  const handleScroll = (e: React.MouseEvent<HTMLAnchorElement>, id: string) => {
    if (!isHome) return;
    e.preventDefault();
    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
      window.history.pushState(null, "", `#${id}`);
      setActiveSection(id);
      setIsMobileMenuOpen(false);
    }
  };

  return (
    <header className="fixed top-0 left-0 right-0 h-16 border-b border-brand-border bg-brand-black/85 backdrop-blur-md z-50 flex items-center px-4 md:px-6 lg:px-12">
      <Link 
        href={isHome ? "#hero" : "/"} 
        onClick={isHome ? (e) => handleScroll(e, "hero") : undefined} 
        className="flex items-center gap-3 mr-auto transition-opacity hover:opacity-80"
      >
        <div className="w-8 h-8 rounded-lg bg-brand-white flex items-center justify-center">
          <Server size={18} className="text-brand-black" />
        </div>
        <span className="font-semibold text-brand-white text-lg tracking-tight">
          Trae<span className="text-brand-gray font-medium">Proxy</span>
        </span>
      </Link>
      
      {/* Desktop Navigation */}
      <nav className="hidden md:flex items-center gap-8 text-sm font-medium">
        {isHome ? (
          <>
            <a 
              href="#hero" 
              onClick={(e) => handleScroll(e, "hero")}
              className={`transition-colors ${activeSection === "hero" ? "text-brand-white" : "text-brand-gray hover:text-brand-white"}`}
            >
              首页
            </a>
            <a 
              href="#why" 
              onClick={(e) => handleScroll(e, "why")}
              className={`transition-colors ${activeSection === "why" ? "text-brand-white" : "text-brand-gray hover:text-brand-white"}`}
            >
              核心优势
            </a>
            <a 
              href="#features" 
              onClick={(e) => handleScroll(e, "features")}
              className={`transition-colors ${activeSection === "features" ? "text-brand-white" : "text-brand-gray hover:text-brand-white"}`}
            >
              工作原理
            </a>
            <a 
              href="#guide" 
              onClick={(e) => handleScroll(e, "guide")}
              className={`transition-colors ${activeSection === "guide" ? "text-brand-white" : "text-brand-gray hover:text-brand-white"}`}
            >
              配置指南
            </a>
          </>
        ) : (
          <span className="text-brand-white transition-opacity">
            使用指南
          </span>
        )}
      </nav>

      {/* Desktop Actions */}
      <div className="hidden md:flex ml-6 items-center">
        <Link 
          href="/get-token" 
          onClick={() => setIsMobileMenuOpen(false)}
          className="flex items-center gap-2 bg-brand-white text-brand-black px-4 py-2.5 rounded-lg text-sm font-semibold transition-all shadow-[0_0_15px_rgba(255,255,255,0.1)] hover:shadow-[0_0_25px_rgba(255,255,255,0.2)]"
        >
          <Key size={16} />
          获取专属凭证
        </Link>
      </div>

      {/* Mobile Menu Toggle Button */}
      <button 
        className="md:hidden ml-auto text-brand-white p-2 -mr-2 focus:outline-none" 
        onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
        aria-label="Toggle menu"
      >
        {isMobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
      </button>

      {/* Mobile Fullscreen Menu */}
      {isMobileMenuOpen && (
        <div className="absolute top-16 left-0 right-0 h-[calc(100vh-4rem)] bg-brand-black/95 backdrop-blur-xl border-b border-brand-border md:hidden flex flex-col pt-8 px-6 pb-8 overflow-y-auto">
          <nav className="flex flex-col gap-6 text-base font-medium w-full mt-4">
            {isHome ? (
              <>
                <a 
                  href="#hero" 
                  onClick={(e) => handleScroll(e, "hero")}
                  className={`w-full text-left py-2 transition-colors ${activeSection === "hero" ? "text-brand-white" : "text-brand-gray hover:text-brand-white"}`}
                >
                  首页
                </a>
                <a 
                  href="#why" 
                  onClick={(e) => handleScroll(e, "why")}
                  className={`w-full text-left py-2 transition-colors ${activeSection === "why" ? "text-brand-white" : "text-brand-gray hover:text-brand-white"}`}
                >
                  核心优势
                </a>
                <a 
                  href="#features" 
                  onClick={(e) => handleScroll(e, "features")}
                  className={`w-full text-left py-2 transition-colors ${activeSection === "features" ? "text-brand-white" : "text-brand-gray hover:text-brand-white"}`}
                >
                  工作原理
                </a>
                <a 
                  href="#guide" 
                  onClick={(e) => handleScroll(e, "guide")}
                  className={`w-full text-left py-2 transition-colors ${activeSection === "guide" ? "text-brand-white" : "text-brand-gray hover:text-brand-white"}`}
                >
                  配置指南
                </a>
              </>
            ) : (
              <span className="w-full text-left py-2 text-brand-white">
                使用指南
              </span>
            )}
          </nav>
          
          <div className="mt-auto pt-8 w-full border-t border-brand-border mt-12">
            <Link 
              href="/get-token" 
              onClick={() => setIsMobileMenuOpen(false)}
              className="flex justify-center items-center gap-2 bg-brand-white text-brand-black px-6 py-4 rounded-xl text-base font-semibold w-full transition-opacity hover:opacity-90 shadow-lg"
            >
              <Key size={18} />
              获取极速专属凭证
            </Link>
          </div>
        </div>
      )}
    </header>
  );
}
