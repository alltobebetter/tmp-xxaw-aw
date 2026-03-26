import { ArrowLeft, BookOpen, DownloadCloud, Key, Rocket, Terminal, Apple, Monitor, AlertTriangle } from "lucide-react";
import Link from "next/link";

export const metadata = {
  title: "使用教程 - TraeProxy",
  description: "如何配置并使用 TraeProxy 无缝接管本地大模型请求",
};

export default function HowToUse() {
  return (
    <div className="flex-1 flex flex-col items-center px-4 md:px-6 lg:px-12 py-16 lg:py-24 font-[family-name:var(--font-sans)] w-full">
      <main className="w-full max-w-3xl text-brand-light">
        
        <Link href="/" className="inline-flex items-center gap-2 text-brand-gray hover:text-brand-white transition-colors mb-12 text-sm font-medium">
          <ArrowLeft size={16} /> 返回首页
        </Link>

        <div className="flex items-center gap-4 mb-6">
          <div className="w-12 h-12 rounded-xl bg-brand-dark border border-brand-border flex items-center justify-center">
            <BookOpen className="text-brand-white" size={24} />
          </div>
          <h1 className="text-4xl sm:text-5xl font-bold tracking-tight text-brand-white">官方配置手册</h1>
        </div>
        <p className="text-lg text-brand-gray leading-relaxed mb-16 pb-12 border-b border-brand-border">
          只需简单三步，即可让 TraeProxy 安全地接管客户端内置大模型的死锁请求。本手册将全面指导您完成闭环体验。
        </p>

        {/* Content Body */}
        <div className="space-y-16">
          <section>
            <div className="flex items-center gap-3 mb-6">
              <Key className="text-brand-white" size={20} />
              <h2 className="text-2xl font-bold text-brand-white">Step 1. 填入自定义反代目标</h2>
            </div>
            <div className="text-brand-gray leading-relaxed space-y-4">
              <p>打开 TraeProxy 客户端，首先在"<strong>API 真实目标源</strong>"输入框中，填入你的第三方 API 流转地址。</p>
              <div className="bg-brand-dark border border-brand-border rounded-lg p-5 my-4 font-mono text-sm text-brand-light">
                <span className="text-brand-white">正确示例：</span><br/>
                <span className="text-green-500 mt-2 block">https://api.openai.com</span>
                <span className="text-green-500 block">https://api.deepseek.com</span>
                <span className="text-green-500 block">https://your-custom-proxy.com</span>
                <br/>
                <span className="text-brand-white">错误示例（切勿携带路径）：</span><br/>
                <span className="text-red-500 mt-2 block">https://api.openai.com/v1</span>
                <span className="text-red-500 block">https://api.openai.com/v1/chat/completions</span>
              </div>
              <p>同时，你可以自定义一个"<strong>本地监听服务端口</strong>"，通常保持默认的 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1.5 py-0.5 rounded text-brand-white font-mono mx-1">8866</code> 即可。如果你本地凑巧被占用了，换一个四位数未绑定端口即可。</p>
            </div>
          </section>

          <section>
            <div className="flex items-center gap-3 mb-6">
              <DownloadCloud className="text-brand-white" size={20} />
              <h2 className="text-2xl font-bold text-brand-white">Step 2. 授权安装系统信任证书</h2>
            </div>
            <div className="text-brand-gray leading-relaxed space-y-4">
              <p>这是整个软件的核心引擎：为了能够在本地破解客户端极其强硬的 HTTPS 加密访问限制，你必须允许 TraeProxy 在你的系统中安装一个仅供局部验证使用的、动态生成的且极简的 Root CA 根证书。</p>
              
              {/* Windows Instructions */}
              <div className="mt-6 p-6 border border-brand-border rounded-xl">
                <div className="flex items-center gap-2 mb-4">
                  <Monitor size={18} className="text-brand-white" />
                  <h3 className="text-lg font-semibold text-brand-white">Windows 操作流程</h3>
                </div>
                <ul className="list-decimal pl-5 space-y-2 text-brand-white/80">
                  <li>点击客户端 HTTPS 证书区域的 <strong className="text-brand-white">"一键激活系统信任"</strong> 按钮。</li>
                  <li>此时，Windows 会弹出一个<strong>用户账户控制 (UAC) 提权窗口</strong>请求管理员权限，请点击"是"。</li>
                  <li>紧接着，系统底层会触发一个 <strong className="text-brand-white">"安全警告"</strong>，请点击"是"允许强行写入系统仓库。</li>
                  <li>当界面上的盾牌指示灯变成 <strong className="text-green-400">绿色</strong>，并显示"已安全信托给系统"，说明底层解密引擎已就绪。</li>
                </ul>
              </div>

              {/* macOS Instructions */}
              <div className="mt-4 p-6 border border-brand-border rounded-xl">
                <div className="flex items-center gap-2 mb-4">
                  <Apple size={18} className="text-brand-white" />
                  <h3 className="text-lg font-semibold text-brand-white">macOS 操作流程</h3>
                </div>
                <ul className="list-decimal pl-5 space-y-2 text-brand-white/80">
                  <li>同样点击 <strong className="text-brand-white">"一键激活系统信任"</strong> 按钮。</li>
                  <li>macOS 会弹出一个<strong>系统密码输入框</strong>（类似 Windows 的 UAC），输入你的 Mac 登录密码并确认。</li>
                  <li>当界面上的盾牌指示灯变成 <strong className="text-green-400">绿色</strong>，说明证书已写入 macOS 钥匙串并完成信任。</li>
                </ul>
              </div>
            </div>
          </section>

          <section>
            <div className="flex items-center gap-3 mb-6">
              <Rocket className="text-brand-white" size={20} />
              <h2 className="text-2xl font-bold text-brand-white">Step 3. 进程级无污染注入拉起</h2>
            </div>
            <div className="text-brand-gray leading-relaxed space-y-4">
              <p>与其他任何代理接管软件最大的不同点在于，你需要<strong>强制在此客户端内唤起目标编辑器</strong>。这样才能保证彻底的"单兵局部注入防溢出"，不污染全局网卡。</p>
              <ul className="list-decimal pl-5 space-y-2 mt-4 text-brand-white/80">
                <li>确保点击了软件极上端那一颗硕大的全区开关，保持绿灯在"运行中"状态。</li>
                <li>
                  在中下部的 <strong>"软件路径"</strong> 选择框内，点击右边的小文件夹图标，找到你电脑中编辑器的原始启动文件。
                  <div className="mt-3 grid grid-cols-1 md:grid-cols-2 gap-3">
                    <div className="p-3 bg-brand-dark border border-brand-border rounded-lg">
                      <div className="flex items-center gap-2 mb-2">
                        <Monitor size={14} className="text-brand-gray" />
                        <span className="text-brand-white text-sm font-medium">Windows</span>
                      </div>
                      <p className="text-xs text-brand-gray">选择 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1 py-0.5 rounded text-brand-white font-mono">Trae.exe</code> 或 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1 py-0.5 rounded text-brand-white font-mono">Trae CN.exe</code></p>
                    </div>
                    <div className="p-3 bg-brand-dark border border-brand-border rounded-lg">
                      <div className="flex items-center gap-2 mb-2">
                        <Apple size={14} className="text-brand-gray" />
                        <span className="text-brand-white text-sm font-medium">macOS</span>
                      </div>
                      <p className="text-xs text-brand-gray">选择 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1 py-0.5 rounded text-brand-white font-mono">Trae.app</code> 或 <code className="bg-brand-gray/20 border border-brand-gray/30 px-1 py-0.5 rounded text-brand-white font-mono">Trae CN.app</code></p>
                    </div>
                  </div>
                </li>
                <li>最后，点击右侧的 <strong className="text-brand-white">"拉起"</strong> 。</li>
              </ul>
              <div className="mt-6 p-5 border border-brand-border bg-brand-dark/40 rounded-xl flex gap-4 items-start">
                <Terminal className="text-brand-white shrink-0" size={20} />
                <p className="text-sm leading-relaxed text-brand-light">大功告成！成功极速拉起后，虽然编辑器表面看起来没有任何异样，但其实它已经被悄悄挂载了局部代理参数。你在里面的任何原生对话请求，都会被你电脑里的 TraeProxy 瞬间截获、改写头寸，并畅通无阻地重定向到你输入的目标源中去！</p>
              </div>
            </div>
          </section>

          {/* macOS First-time Setup Note */}
          <section className="bg-brand-dark/20 p-8 rounded-2xl border border-brand-border">
            <div className="flex items-center gap-3 mb-4">
              <Apple size={20} className="text-brand-white" />
              <h3 className="text-xl font-bold text-brand-white">macOS 首次使用须知</h3>
            </div>
            <div className="text-brand-gray text-sm leading-relaxed space-y-3">
              <p>由于 TraeProxy 目前未经过 Apple 开发者签名，macOS 的 Gatekeeper 会在首次打开时进行拦截。请使用以下任一方法解除限制：</p>
              <div className="p-4 bg-black/40 border border-brand-border rounded-xl space-y-3 mt-3">
                <p className="text-brand-light"><strong>方法一（推荐）：</strong>在 Finder 中找到 TraeProxy.app，<strong className="text-brand-white">右键点击</strong> → 选择「打开」→ 在弹出的对话框中再次点击「打开」。</p>
                <p className="text-brand-light"><strong>方法二（终端）：</strong>打开终端执行：<code className="bg-brand-gray/20 border border-brand-gray/30 px-1.5 py-0.5 rounded font-mono text-brand-white ml-1">xattr -cr /Applications/TraeProxy.app</code></p>
              </div>
              <p className="text-brand-gray text-xs mt-2">只需操作一次，之后可正常双击启动。此限制仅因未购买 Apple 开发者签名（$99/年），与软件安全性无关。</p>
            </div>
          </section>

          <section className="bg-brand-dark/20 p-8 rounded-2xl border border-brand-border">
            <h3 className="text-xl font-bold text-brand-white mb-4">用完即抛，安全退出</h3>
            <p className="text-brand-gray text-sm leading-relaxed mb-4">
              不用时，直接点击客户端右上角的红叉将其关闭，所有的内存和端口占用会瞬时"魂归天际"。
            </p>
            <p className="text-brand-gray text-sm leading-relaxed">
              如果你不再使用本软件，点击界面最底部的"一键清除数据"，可以在完全不误伤系统底层的情况下，将所有本地配置清空。另外若是想彻底移除那个已部署的系统根证书，请点击证书区域上方独立的垃圾桶卸载图标，即刻彻底连根拔除。
            </p>
          </section>

        </div>
      </main>
    </div>
  );
}
