"use client";

import { useState, useEffect, useRef } from "react";
import { Key, Copy, CheckCircle2, ChevronLeft } from "lucide-react";
import Link from "next/link";
import Script from "next/script";

declare global {
  interface Window {
    grecaptcha: any;
    onRecaptchaLoad: () => void;
  }
}

export default function GetTokenPage() {
  const [loading, setLoading] = useState(false);
  const [token, setToken] = useState("");
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState("");
  const [recaptchaToken, setRecaptchaToken] = useState("");
  const recaptchaRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const renderRecaptcha = () => {
      if (window.grecaptcha && window.grecaptcha.render && recaptchaRef.current) {
        if (!recaptchaRef.current.hasChildNodes()) {
          try {
            window.grecaptcha.render(recaptchaRef.current, {
              sitekey: process.env.NEXT_PUBLIC_RECAPTCHA_SITE_KEY || "6Lcz5JYsAAAAAE23XZtkqlJlijfacJmfvIq-DKGt",
              theme: "dark",
              callback: (t: string) => setRecaptchaToken(t),
              "expired-callback": () => setRecaptchaToken(""),
            });
          } catch (e) {
            console.error("Recaptcha render error", e);
          }
        }
      }
    };

    if (window.grecaptcha) {
      renderRecaptcha();
    } else {
      window.onRecaptchaLoad = renderRecaptcha;
    }
  }, []);

  const handleGenerate = async () => {
    if (!recaptchaToken) {
      setError("请先完成上方的人机验证");
      return;
    }
    setLoading(true);
    setError("");
    try {
      const res = await fetch("/api/auth/generate", { 
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ recaptchaToken })
      });
      const data = await res.json();
      if (data.success && data.data.token) {
        setToken(data.data.token);
      } else {
        setError(data.error || "Failed to generate token");
      }
    } catch (err) {
      setError("Network error occurred");
    } finally {
      setLoading(false);
    }
  };

  const handleCopy = () => {
    if (!token) return;
    navigator.clipboard.writeText(token);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="flex-1 flex flex-col items-center justify-center p-6 lg:p-12 font-[family-name:var(--font-sans)] w-full min-h-[calc(100vh-4rem)] relative">
      <div className="absolute top-8 left-8 sm:top-12 sm:left-12 z-10">
        <Link href="/" className="flex items-center gap-2 text-sm font-medium text-brand-gray hover:text-brand-white transition-colors">
          <ChevronLeft size={16} /> 返回首页
        </Link>
      </div>

      <div className="w-full max-w-3xl flex flex-col items-center text-center mt-[-8vh]">
        {!token ? (
          <>
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-brand-border bg-brand-dark text-xs text-brand-gray uppercase tracking-wider mb-6">
              <Key size={14} /> 系统级授权凭证
            </div>
            
            <h1 className="text-4xl sm:text-[3.5rem] font-bold tracking-tight text-brand-white leading-tight mb-6">
              获取专属鉴权密钥
            </h1>
            <p className="text-lg sm:text-xl text-brand-gray max-w-2xl leading-relaxed mb-12">
              此密钥是激活 TraeProxy 桌面端的唯一核验凭证。<br className="hidden sm:block" />请妥善保管，切勿泄露给第三方。
            </p>

            <div className="w-full max-w-sm flex flex-col items-center">
              <div className="mb-6 w-full flex justify-center">
                <div className="w-[300px] h-[74px] relative overflow-hidden rounded-[4px] border border-brand-border bg-[#222]">
                  <div className="absolute top-[-2px] left-[-2px]">
                    <div ref={recaptchaRef}></div>
                  </div>
                </div>
              </div>
              <div className="relative w-full">
                <span className="absolute -top-3 -right-3 z-10 bg-red-500 text-white text-[11px] font-bold px-2 py-0.5 rounded-full shadow-lg transform rotate-12">限时免费</span>
                <button
                  onClick={handleGenerate}
                  disabled={loading || !recaptchaToken}
                  className="w-full bg-brand-white text-brand-black hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed font-semibold py-4 rounded-lg transition-opacity flex items-center justify-center gap-2 text-[15px]"
                >
                  {loading ? (
                    <span className="animate-pulse">密钥生成中...</span>
                  ) : (
                    "生成我的Token"
                  )}
                </button>
              </div>
            {error && <p className="text-red-500 text-sm mt-4">{error}</p>}
            </div>
          </>
        ) : (
          <div className="w-full flex flex-col items-center">
            <div className="w-full py-8 flex flex-col items-center mb-4">
              <div className="text-sm text-brand-gray mb-4 uppercase tracking-widest font-semibold select-none">Your Access Token</div>
              <div className="text-4xl sm:text-5xl font-mono text-brand-white tracking-wider font-bold break-all selection:bg-brand-gray">
                {token}
              </div>
            </div>
            
            <button
              onClick={handleCopy}
              className="w-full max-w-sm bg-brand-white text-brand-black hover:opacity-90 font-semibold py-4 rounded-lg transition-opacity flex items-center justify-center gap-2 text-[15px]"
            >
              {copied ? (
                <><CheckCircle2 size={18} className="text-green-500" /> 已安全复制</>
              ) : (
                <><Copy size={18} /> 一键复制到剪贴板</>
              )}
            </button>
            <p className="text-brand-gray text-xs mt-6 px-4">凭证已同步至云端。请打开 TraeProxy 桌面客户端填入本凭证进行激活。</p>
          </div>
        )}
      </div>
      <Script 
        src="https://www.recaptcha.net/recaptcha/api.js?onload=onRecaptchaLoad&render=explicit" 
        strategy="lazyOnload" 
      />
    </div>
  );
}
