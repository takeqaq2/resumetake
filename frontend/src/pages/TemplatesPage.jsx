import React from 'react'
import { Link } from 'react-router-dom'
import { Helmet } from 'react-helmet-async'

const templates = [
  { id: 'professional', name: '专业商务', desc: '适合传统行业和商务岗位', gradient: 'from-blue-500 to-blue-600' },
  { id: 'modern', name: '现代简约', desc: '适合互联网和科技行业', gradient: 'from-violet-500 to-purple-600' },
  { id: 'creative', name: '创意设计', desc: '适合设计和创意岗位', gradient: 'from-pink-500 to-rose-600' },
  { id: 'academic', name: '学术科研', desc: '适合教育和研究岗位', gradient: 'from-emerald-500 to-green-600' },
  { id: 'executive', name: '高管专用', desc: '适合高级管理岗位', gradient: 'from-slate-700 to-slate-900' },
  { id: 'minimal', name: '极简风格', desc: '简洁大方，通用性强', gradient: 'from-amber-500 to-orange-600' },
]

function TemplatesPage() {
  return (
    <>
      <Helmet>
        <title>简历模板 - ResumeTake | 免费专业简历模板</title>
        <meta name="description" content="多种专业简历模板，覆盖互联网、金融、教育等行业，免费使用。" />
      </Helmet>

      <div className="max-w-6xl mx-auto px-4 py-12">
        <div className="text-center mb-12">
          <h1 className="section-title mb-3">专业简历模板</h1>
          <p className="section-subtitle">选择适合你行业的模板，一键套用</p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {templates.map(t => (
            <Link key={t.id} to={`/editor?template=${t.id}`} className="group card-hover">
              <div className={`h-40 bg-gradient-to-br ${t.gradient} rounded-xl mb-4 flex items-center justify-center group-hover:scale-[1.02] transition-transform`}>
                <span className="text-white/20 text-6xl font-bold">Aa</span>
              </div>
              <h3 className="font-semibold text-slate-900 dark:text-white mb-1">{t.name}</h3>
              <p className="text-sm text-slate-500 dark:text-slate-400">{t.desc}</p>
            </Link>
          ))}
        </div>
      </div>
    </>
  )
}

export default TemplatesPage
