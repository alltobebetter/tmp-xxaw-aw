import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Navbar from "@/components/Navbar";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "TraeProxy - 极简本地代理网关",
  description: "专为无缝接管大模型请求而设计，一次配置，全局静默运行。",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="zh-CN"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased`}
      data-scroll-behavior="smooth"
    >
      <body className="min-h-full flex flex-col bg-brand-black text-brand-light">
        <Navbar />
        <div className="flex-1 mt-16 flex flex-col">
          {children}
        </div>
        <footer className="py-8 text-center text-sm text-brand-gray/60 border-t border-brand-border mt-auto">
          &copy; {new Date().getFullYear()} TraeProxy. Built With AgentLab.
        </footer>
      </body>
    </html>
  );
}
