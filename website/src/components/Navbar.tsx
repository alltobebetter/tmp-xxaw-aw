"use client";

import { useEffect, useState } from "react";
import { Server } from "lucide-react";
import { usePathname } from "next/navigation";
import Link from "next/link";

export default function Navbar() {
  const [activeSection, setActiveSection] = useState("hero");
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

  const handleScroll = (e: React.MouseEvent<HTMLAnchorElement>, id: string) => {
    if (!isHome) return;
    e.preventDefault();
    const element = document.getElementById(id);
    if (element) {
      element.scrollIntoView({ behavior: "smooth" });
      window.history.pushState(null, "", `#${id}`);
    }
  };

  return (
    <header className="fixed top-0 left-0 right-0 h-16 border-b border-brand-border bg-brand-black/85 backdrop-blur-md z-50 flex items-center px-6 lg:px-12">
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
      
      <nav className="flex items-center gap-8 text-sm font-medium">
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
    </header>
  );
}
