"use client";

import { ShieldCheck, Zap, ServerCrash, Network, Cpu, Key, DownloadCloud, Rocket, Download, BookOpen, Apple, Monitor, X, CheckCircle2, AlertTriangle, FolderOpen } from "lucide-react";
import { useState, useEffect, useRef, useCallback } from "react";
import Link from "next/link";

function MacInstallModal({ open, onClose, downloadUrl }: { open: boolean; onClose: () => void; downloadUrl: string }) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm" onClick={onClose}>
      <div className="bg-brand-dark border border-brand-border rounded-2xl max-w-lg w-[90%] p-8 relative shadow-2xl" onClick={e => e.stopPropagation()}>
        <button onClick={onClose} className="absolute top-4 right-4 text-brand-gray hover:text-brand-white transition-colors">
          <X size={20} />
        </button>
        
        <div className="flex items-center gap-3 mb-6">
          <Apple size={28} className="text-brand-white" />
          <h2 className="text-2xl font-bold text-brand-white">macOS 安装引导</h2>
        </div>

        <div className="space-y-5 text-sm">
          <div className="flex gap-3 items-start">
            <div className="w-6 h-6 rounded-full bg-brand-border flex items-center justify-center shrink-0 mt-0.5 text-xs font-bold text-brand-white">1</div>
            <div>
              <p className="text-brand-white font-medium">下载并解压</p>
              <p className="text-brand-gray mt-1">下载 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1.5 py-0.5 rounded font-mono text-brand-white mx-0.5">.zip</code> 文件后，双击解压得到 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1.5 py-0.5 rounded font-mono text-brand-white mx-0.5">TraeProxy.app</code></p>
            </div>
          </div>

          <div className="flex gap-3 items-start">
            <div className="w-6 h-6 rounded-full bg-brand-border flex items-center justify-center shrink-0 mt-0.5 text-xs font-bold text-brand-white">2</div>
            <div>
              <p className="text-brand-white font-medium">拖入 Applications</p>
              <p className="text-brand-gray mt-1">将 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1.5 py-0.5 rounded font-mono text-brand-white mx-0.5">TraeProxy.app</code> 拖入 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1.5 py-0.5 rounded font-mono text-brand-white mx-0.5">/Applications</code> 文件夹（也可以放在其他位置）</p>
            </div>
          </div>

          <div className="flex gap-3 items-start">
            <div className="w-6 h-6 rounded-full bg-brand-border flex items-center justify-center shrink-0 mt-0.5 text-xs font-bold text-brand-white">3</div>
            <div>
              <p className="text-brand-white font-medium">绕过 Gatekeeper 安全检查</p>
              <p className="text-brand-gray mt-1">由于没有 Apple 开发者签名，首次打开会被拦截。</p>
              <div className="mt-3 p-4 bg-black/40 border border-brand-border rounded-xl space-y-2">
                <div className="flex items-start gap-2">
                  <CheckCircle2 size={14} className="text-green-500 shrink-0 mt-0.5" />
                  <p className="text-brand-light"><strong>方法一（推荐）：</strong>右键点击 TraeProxy.app → 选择「打开」→ 弹窗中点「打开」</p>
                </div>
                <div className="flex items-start gap-2">
                  <CheckCircle2 size={14} className="text-green-500 shrink-0 mt-0.5" />
                  <p className="text-brand-light"><strong>方法二（终端）：</strong><code className="bg-brand-gray/20 border border-brand-gray/30 px-1.5 py-0.5 rounded font-mono text-brand-white">xattr -cr /Applications/TraeProxy.app</code></p>
                </div>
              </div>
              <p className="text-brand-gray mt-2 text-xs">只需操作一次，之后可正常双击启动。</p>
            </div>
          </div>

          <div className="flex gap-3 items-start">
            <div className="w-6 h-6 rounded-full bg-brand-border flex items-center justify-center shrink-0 mt-0.5 text-xs font-bold text-brand-white">4</div>
            <div>
              <p className="text-brand-white font-medium">安装证书时</p>
              <p className="text-brand-gray mt-1">点击「一键激活系统信任」后，macOS 会弹出密码输入框（类似 Windows 的 UAC），输入你的登录密码即可完成安装。</p>
            </div>
          </div>
        </div>

        <a
          href={downloadUrl}
          className="mt-8 flex items-center justify-center gap-2 w-full bg-brand-white text-brand-black py-3 rounded-lg text-sm font-semibold transition-opacity hover:opacity-90"
          onClick={onClose}
        >
          <Download size={16} />确认下载 macOS 版本
        </a>
      </div>
    </div>
  );
}

export default function Home() {
  const version = '1.2.0';
  const windowsUrl = `https://public.agentlab.click/TraeProxy_Setup_v${version}.exe`;
  const macUrl = `https://public.agentlab.click/TraeProxy_macOS_v${version}.zip`;
  const [showMacModal, setShowMacModal] = useState(false);
  const macBtnRef = useRef<HTMLButtonElement>(null);
  const [, forceUpdate] = useState(0);

  // Fix: bfcache restore breaks React's synthetic event system.
  // Use native DOM event listener as fallback and force re-render on pageshow.
  useEffect(() => {
    const handlePageShow = (e: PageTransitionEvent) => {
      if (e.persisted) {
        forceUpdate(n => n + 1);
      }
    };
    window.addEventListener('pageshow', handlePageShow);
    return () => window.removeEventListener('pageshow', handlePageShow);
  }, []);

  useEffect(() => {
    const btn = macBtnRef.current;
    if (!btn) return;
    const handler = () => setShowMacModal(true);
    btn.addEventListener('click', handler);
    return () => btn.removeEventListener('click', handler);
  });

  return (
    <div className="flex-1 flex flex-col items-center px-4 md:px-6 lg:px-12 font-[family-name:var(--font-sans)] w-full">
      
      <MacInstallModal open={showMacModal} onClose={() => setShowMacModal(false)} downloadUrl={macUrl} />

      {/* === HERO SECTION === */}
      <section id="hero" className="flex flex-col items-center justify-center gap-10 max-w-4xl text-center w-full min-h-[calc(100vh-4rem)] scroll-mt-16 pb-16">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-brand-border bg-brand-dark text-xs text-brand-gray uppercase tracking-wider mb-2">
          <ServerCrash size={14} /> 解决大模型连通性障碍
        </div>
        <div className="flex flex-col items-center gap-6">
          <h1 className="text-5xl sm:text-[5.5rem] font-bold tracking-tight text-brand-white leading-tight">
            无缝接管，突破封锁。
          </h1>
          <p className="text-lg sm:text-2xl text-brand-gray max-w-3xl leading-relaxed mt-4">
            破解内置大模型 API 锁定，无需污染全局网络，彻底实现「自定义 BaseURL」自由。
          </p>
        </div>
        <div className="flex flex-col sm:flex-row gap-4 mt-8 w-full sm:w-auto">
          <a className="flex items-center justify-center gap-2 bg-brand-white text-brand-black px-10 py-3.5 rounded-lg text-sm font-semibold transition-opacity hover:opacity-90 active:opacity-100" href={windowsUrl}>
            <Monitor size={18} />下载 Windows 版本
          </a>
          <button
            ref={macBtnRef}
            type="button"
            className="flex items-center justify-center gap-2 bg-brand-white text-brand-black px-10 py-3.5 rounded-lg text-sm font-semibold transition-opacity hover:opacity-90 active:opacity-100 cursor-pointer w-full sm:w-auto"
          >
            <Apple size={18} />下载 macOS 版本
          </button>
          <Link className="flex items-center justify-center gap-2 bg-brand-black text-brand-light border border-brand-border px-10 py-3.5 rounded-lg text-sm font-semibold transition-colors hover:bg-brand-dark" href="/how-to-use">
            <BookOpen size={18} />使用指南
          </Link>
        </div>
      </section>

      {/* === OVERVIEW SECTION === */}
      <section id="why" className="w-full max-w-4xl py-24 flex flex-col items-center border-t border-brand-border scroll-mt-24">
        <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-brand-white mb-6 w-full text-left">为什么需要 TraeProxy？</h2>
        <div className="text-lg text-brand-gray mb-16 max-w-4xl leading-relaxed text-left w-full space-y-4">
          <p>当代码编辑器（如 Trae）无法直接在界面修改官方内置接口时，这把钥匙能从系统底层强制拦截并重写向 API 发出的模型对话请求。</p>
          <p>作为本地透明网关，它无需建立全局代理，不干扰任何无关的日常上网。仅仅通过极其克制的"单兵注入"，就能让你无缝接管流量，自由接入任意第三方便宜的兼容模型资源。</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 w-full text-left">
          <div className="p-8 bg-brand-dark border border-brand-border rounded-xl">
            <ShieldCheck className="text-brand-light mb-6" size={28} />
            <h3 className="text-xl font-semibold text-brand-white mb-3">HTTPS 底层拦截</h3>
            <p className="text-sm text-brand-gray leading-relaxed">内置 Root CA 一键生成与卸载机制。直接在系统底层实现安全的中间人解密，稳定截获加密大模型请求。</p>
          </div>
          <div className="p-8 bg-brand-dark border border-brand-border rounded-xl">
            <Zap className="text-brand-light mb-6" size={28} />
            <h3 className="text-xl font-semibold text-brand-white mb-3">孤岛级进程注入</h3>
            <p className="text-sm text-brand-gray leading-relaxed">独创"单兵拉起"模式。不改变任何系统全局代理，不影响浏览器日常上网，仅给目标编辑器注入代理变量。</p>
          </div>
          <div className="p-8 bg-brand-dark border border-brand-border rounded-xl">
            <ServerCrash className="text-brand-light mb-6" size={28} />
            <h3 className="text-xl font-semibold text-brand-white mb-3">零开销理念</h3>
            <p className="text-sm text-brand-gray leading-relaxed">基于纯底层原生 API 编译构建，卸下了沉重的浏览器内核负载。内存耗用微小，毫无系统后台自启与驻留残留。</p>
          </div>
        </div>
      </section>

      {/* === FEATURES SECTION === */}
      <section id="features" className="w-full max-w-4xl py-24 scroll-mt-24 border-t border-brand-border">
        <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-brand-white mb-6">工作原理机制</h2>
        <p className="text-lg text-brand-gray mb-16 max-w-2xl leading-relaxed">TraeProxy 不仅仅是一个代理服务器，它是基于"零污染"哲学构建的无声拦截网关。它是如何做到隐秘而高效的？</p>
        <div className="space-y-12">
          <div className="flex flex-col md:flex-row gap-8 items-start">
            <div className="w-12 h-12 shrink-0 rounded-xl bg-brand-dark border border-brand-border flex items-center justify-center mt-1"><ShieldCheck className="text-brand-light" size={24} /></div>
            <div>
              <h3 className="text-2xl font-semibold text-brand-white mb-4">MITM 中间人拦截技术</h3>
              <p className="text-brand-gray leading-relaxed mb-4">大模型客户端通常内置了强制的 HTTPS 加密访问（例如锁定死请求 api.openai.com）。为了在本地截获并重定向这些请求，TraeProxy 采用 MITM (Man-in-the-Middle) 技术。</p>
              <ul className="list-disc pl-5 space-y-2 text-brand-gray text-sm">
                <li>软件首次运行时，将自动在内存中为您签发独一无二的 Root CA 根证书。</li>
                <li>通过一键操作，利用系统原生提权机制将该证书写入 Windows 系统的"受信任的根证书颁发机构"中。</li>
                <li>TraeProxy 接管端口后，能够解开客户端极其固执的加密封装，提取真实的请求体，并篡改其 Host 将流量运往你指定的反代源。</li>
              </ul>
            </div>
          </div>
          <hr className="border-brand-border" />
          <div className="flex flex-col md:flex-row gap-8 items-start">
            <div className="w-12 h-12 shrink-0 rounded-xl bg-brand-dark border border-brand-border flex items-center justify-center mt-1"><Network className="text-brand-light" size={24} /></div>
            <div>
              <h3 className="text-2xl font-semibold text-brand-white mb-4">进程级别的纯净隔离</h3>
              <p className="text-brand-gray leading-relaxed mb-4">通常的代理软件为了接管软件请求，往往暴力的开启系统级 Tun 网卡或修改系统全局代理环境变量。这直接导致你日常开发浏览其他无关网站也会遭遇异常阻断。</p>
              <ul className="list-disc pl-5 space-y-2 text-brand-gray text-sm">
                <li>TraeProxy 提出"单兵拉起"模式。它绝不仅是一个服务端，更是一个启动器。</li>
                <li>用户通过客户端界面的"拉起"按钮唤醒目标编辑器，软件底层仅对被拉起的那个独立子进程强制注入局部代理参数。</li>
                <li>这种进程隔离级别的拦截，确保了你的操作系统网络环境 100% 不受波及。</li>
              </ul>
            </div>
          </div>
          <hr className="border-brand-border" />
          <div className="flex flex-col md:flex-row gap-8 items-start">
            <div className="w-12 h-12 shrink-0 rounded-xl bg-brand-dark border border-brand-border flex items-center justify-center mt-1"><Cpu className="text-brand-light" size={24} /></div>
            <div>
              <h3 className="text-2xl font-semibold text-brand-white mb-4">用完即焚的极客克制</h3>
              <p className="text-brand-gray leading-relaxed mb-4">没有后台自启动，没有托盘驻留"幽灵"进程。所有的设置参数在点击关闭的那一瞬间，所有相关的代理服务与守护线程全部硬销毁，瞬间归还你的硬件与端口资源。这既是极简主义，也是安全理念。</p>
            </div>
          </div>
        </div>
      </section>

      {/* === GUIDE SECTION === */}
      <section id="guide" className="w-full max-w-4xl py-24 scroll-mt-24 border-t border-brand-border">
        <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-brand-white mb-6">配置与使用指南</h2>
        <p className="text-lg text-brand-gray mb-16 max-w-2xl leading-relaxed">三步设置，即可让您的目标编辑器通过本地代理顺畅接管大模型请求。</p>
        <div className="relative border-l border-brand-border ml-6 pl-10 space-y-16 pb-12">
          <div className="relative">
            <div className="absolute -left-[58px] w-12 h-12 rounded-full bg-brand-black border-2 border-brand-border flex items-center justify-center"><Key className="text-brand-light" size={18} /></div>
            <h3 className="text-xl font-semibold text-brand-white mb-3">第一步：配置反代目标源</h3>
            <p className="text-brand-gray text-sm leading-relaxed mb-3">打开 TraeProxy 客户端，在"路由流转规则"面板中，输入你所购买或部署的第三方大模型 API 直连地址。</p>
            <div className="p-4 bg-brand-dark border border-brand-border rounded-lg text-sm text-brand-gray font-mono mt-4">
              <span className="text-brand-border select-none mr-2">1</span> 例如: <span className="text-brand-white">https://api.openai.com</span><br/>
              <span className="text-brand-border select-none mr-2">2</span> (请注意：不要在该网址后面附带 /v1 等路径后缀)
            </div>
          </div>
          <div className="relative">
            <div className="absolute -left-[58px] w-12 h-12 rounded-full bg-brand-black border-2 border-brand-border flex items-center justify-center"><DownloadCloud className="text-brand-light" size={18} /></div>
            <h3 className="text-xl font-semibold text-brand-white mb-3">第二步：安装系统级信任证书</h3>
            <p className="text-brand-gray text-sm leading-relaxed mb-3">在安全状态区域，如果你看到红色的警示徽章，点击 <b>"一键激活系统信任"</b>。</p>
            <p className="text-brand-gray text-sm leading-relaxed mb-3">Windows 会弹出一个安全提权（UAC）窗口以及一个安全警告，请点击"是"允许安装。此时绿色徽章亮起，表示底层解密引擎已就绪。</p>
          </div>
          <div className="relative">
            <div className="absolute -left-[58px] w-12 h-12 rounded-full bg-brand-black border-2 border-brand-border flex items-center justify-center"><Rocket className="text-brand-light" size={18} /></div>
            <h3 className="text-xl font-semibold text-brand-white mb-3">第三步：唤醒代理进程</h3>
            <p className="text-brand-gray text-sm leading-relaxed mb-3">必须在左侧面板确保基础网络开关已处于"运行中"的状态。接着选择你需要注入的编辑器的主程序执行文件（如 Trae.exe），然后点击右侧的 <b>"拉起"</b> 按钮。</p>
            <p className="text-brand-gray text-sm leading-relaxed mb-3">大功告成。在你的编辑器中直接向原厂模型发起对话，它将被无缝透明传送至你的目标源。</p>
          </div>
        </div>
      </section>
    </div>
  );
}
