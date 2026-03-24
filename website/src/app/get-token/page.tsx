"use client";

import { useState } from "react";
import { Key, Copy, CheckCircle2, ChevronLeft } from "lucide-react";
import Link from "next/link";

export default function GetTokenPage() {
  const [loading, setLoading] = useState(false);
  const [token, setToken] = useState("");
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState("");

  const handleGenerate = async () => {
    setLoading(true);
    setError("");
    try {
      const res = await fetch("/api/auth/generate", { method: "POST" });
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

      <div className="w-full max-w-lg p-10 bg-brand-dark/80 border border-brand-border rounded-2xl flex flex-col items-center text-center shadow-2xl">
        <div className="w-16 h-16 rounded-full bg-brand-black border-2 border-brand-border flex items-center justify-center mb-6 shadow-inner">
          <Key className="text-brand-white" size={24} />
        </div>
        <h1 className="text-3xl sm:text-4xl font-bold text-brand-white mb-4 tracking-tight">获取专属鉴权密钥</h1>
        <p className="text-brand-gray text-sm sm:text-base leading-relaxed mb-10 w-full px-4">
          此密钥是激活 TraeProxy 桌面端的唯一核验凭证。<br className="hidden sm:block" />请妥善保管，切勿泄露给第三方。
        </p>

        {!token ? (
          <div className="w-full">
            <button
              onClick={handleGenerate}
              disabled={loading}
              className="w-full bg-brand-white text-brand-black hover:opacity-90 disabled:opacity-50 disabled:cursor-wait font-semibold py-4 rounded-xl transition-all flex items-center justify-center gap-2 text-base"
            >
              {loading ? (
                <span className="animate-pulse">密钥生成中...</span>
              ) : (
                "免费生成我的 Token"
              )}
            </button>
            {error && <p className="text-red-500 text-sm mt-4 bg-red-500/10 p-3 rounded-lg border border-red-500/20">{error}</p>}
          </div>
        ) : (
          <div className="w-full flex flex-col items-center animate-in fade-in zoom-in-95 duration-500">
            <div className="w-full bg-brand-black border border-brand-border rounded-xl p-8 flex flex-col items-center mb-6 shadow-inner">
              <div className="text-xs text-brand-gray mb-3 uppercase tracking-widest font-semibold">Your Access Token</div>
              <div className="text-3xl sm:text-4xl font-mono text-brand-white tracking-wider font-bold break-all selection:bg-brand-gray">
                {token}
              </div>
            </div>
            
            <button
              onClick={handleCopy}
              className="w-full bg-brand-black border border-brand-border text-brand-white hover:bg-brand-white hover:text-brand-black font-semibold py-4 rounded-xl transition-all flex items-center justify-center gap-2 text-base"
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
    </div>
  );
}
