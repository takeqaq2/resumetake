import React from 'react'
import { Link } from 'react-router-dom'
import { Helmet } from 'react-helmet-async'
import { HiSparkles, HiMagnifyingGlass, HiDocumentCheck, HiArrowRight } from 'react-icons/hi2'

function HomePage() {
  return (
    <>
      <Helmet>
        <title>AI简历优化工具 - ResumeTake | 智能简历制作平台</title>
        <meta name="description" content="AI智能简历优化工具，一键生成专业简历，ATS关键词匹配，提升求职成功率。免费在线简历制作，支持PDF导出。" />
        <meta name="keywords" content="AI简历优化,简历生成器,智能简历制作,ATS简历优化,免费简历工具,在线简历编辑" />
        <link rel="canonical" href="https://resume.takee.top/" />
        <meta property="og:title" content="AI简历优化工具 - ResumeTake" />
        <meta property="og:description" content="AI智能简历优化，一键生成专业简历，ATS关键词匹配。" />
        <meta property="og:url" content="https://resume.takee.top/" />
        <meta property="og:type" content="website" />
        <script type="application/ld+json">
          {JSON.stringify({
            "@context": "https://schema.org",
            "@type": "SoftwareApplication",
            "name": "ResumeTake AI简历优化工具",
            "description": "智能AI简历优化工具，一键生成专业简历，ATS关键词匹配",
            "url": "https://resume.takee.top",
            "applicationCategory": "BusinessApplication",
            "operatingSystem": "Web",
            "offers": { "@type": "Offer", "price": "0", "priceCurrency": "CNY" }
          })}
        </script>
      </Helmet>

      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-blue-50 via-white to-violet-50 dark:from-slate-900 dark:via-slate-900 dark:to-slate-800" />
        <div className="relative max-w-6xl mx-auto px-4 pt-20 pb-24 md:pt-32 md:pb-36">
          <div className="text-center max-w-3xl mx-auto">
            <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 text-sm font-medium mb-6">
              <HiSparkles className="w-4 h-4" />
              AI驱动的智能简历优化
            </div>
            <h1 className="text-4xl md:text-6xl font-bold text-slate-900 dark:text-white mb-6 tracking-tight text-balance">
              让AI帮你打造
              <span className="gradient-text">完美简历</span>
            </h1>
            <p className="text-lg md:text-xl text-slate-600 dark:text-slate-400 mb-10 text-balance">
              上传简历，AI自动优化内容、匹配ATS关键词、提升通过率。
              一键生成专业简历，助你获得理想工作。
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link to="/editor" className="btn-primary text-base py-3 px-8 inline-flex items-center justify-center gap-2">
                立即开始
                <HiArrowRight className="w-5 h-5" />
              </Link>
              <a href="#features" className="btn-secondary text-base py-3 px-8">
                了解更多
              </a>
            </div>
          </div>
        </div>
      </section>

      <section id="features" className="py-24 bg-white dark:bg-slate-900">
        <div className="max-w-6xl mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="section-title mb-4">为什么选择 ResumeTake</h2>
            <p className="section-subtitle">基于先进AI技术，让简历优化变得简单高效</p>
          </div>
          <div className="grid md:grid-cols-3 gap-8">
            {[
              {
                icon: HiSparkles,
                title: 'AI智能优化',
                desc: '自动优化简历内容、用词和格式，突出你的核心竞争力与成就',
                color: 'from-blue-500 to-cyan-500',
              },
              {
                icon: HiMagnifyingGlass,
                title: 'ATS关键词匹配',
                desc: '智能分析目标职位，自动匹配ATS关键词，大幅提升通过率',
                color: 'from-violet-500 to-purple-500',
              },
              {
                icon: HiDocumentCheck,
                title: '专业模板',
                desc: '多种专业模板覆盖各行各业，一键切换风格，导出高质量PDF',
                color: 'from-amber-500 to-orange-500',
              },
            ].map((feature, i) => (
              <div key={i} className="group p-8 rounded-2xl bg-slate-50 dark:bg-slate-800/50 hover:bg-white dark:hover:bg-slate-800 border border-transparent hover:border-slate-200 dark:hover:border-slate-700 hover:shadow-lg transition-all duration-300">
                <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${feature.color} flex items-center justify-center mb-5 group-hover:scale-110 transition-transform`}>
                  <feature.icon className="w-6 h-6 text-white" />
                </div>
                <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-2">{feature.title}</h3>
                <p className="text-slate-600 dark:text-slate-400 leading-relaxed">{feature.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="py-24 bg-slate-50 dark:bg-slate-800/30">
        <div className="max-w-6xl mx-auto px-4">
          <div className="text-center mb-16">
            <h2 className="section-title mb-4">简单三步完成</h2>
            <p className="section-subtitle">无需复杂操作，轻松获得专业简历</p>
          </div>
          <div className="grid md:grid-cols-3 gap-8">
            {[
              { step: '01', title: '上传简历', desc: '支持PDF、Word格式，一键上传' },
              { step: '02', title: 'AI优化', desc: '智能分析并优化简历内容' },
              { step: '03', title: '下载简历', desc: '一键导出专业PDF简历' },
            ].map((item, i) => (
              <div key={i} className="text-center">
                <div className="text-5xl font-bold gradient-text mb-4 opacity-50">{item.step}</div>
                <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-2">{item.title}</h3>
                <p className="text-slate-600 dark:text-slate-400">{item.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      <section className="py-24">
        <div className="max-w-4xl mx-auto px-4 text-center">
          <h2 className="section-title mb-6">开始优化你的简历</h2>
          <p className="section-subtitle mb-10">免费使用AI简历优化工具，提升求职成功率</p>
          <Link to="/editor" className="btn-primary text-base py-3 px-10 inline-flex items-center gap-2">
            免费开始
            <HiArrowRight className="w-5 h-5" />
          </Link>
        </div>
      </section>
    </>
  )
}

export default HomePage
