import React, { useState } from 'react'
import { useParams } from 'react-router-dom'
import { Helmet } from 'react-helmet-async'
import { HiUpload, HiSparkles, HiDocumentArrowDown } from 'react-icons/hi2'
import axios from 'axios'
import { toast } from 'react-toastify'

function EditorPage() {
  const { id } = useParams()
  const [resume, setResume] = useState({ name: '', email: '', phone: '', summary: '', skills: [] })
  const [targetJob, setTargetJob] = useState('')
  const [jobDescription, setJobDescription] = useState('')
  const [isOptimizing, setIsOptimizing] = useState(false)
  const [result, setResult] = useState(null)
  const [tab, setTab] = useState('edit')

  const handleOptimize = async () => {
    setIsOptimizing(true)
    try {
      const res = await axios.post('/api/v1/optimize', { resume_content: resume, target_job: targetJob, job_description: jobDescription })
      setResult(res.data.data)
      toast.success('简历优化完成！')
      setTab('result')
    } catch { toast.error('优化失败，请重试') }
    finally { setIsOptimizing(false) }
  }

  return (
    <>
      <Helmet>
        <title>简历编辑器 - ResumeTake</title>
        <meta name="description" content="使用AI智能优化你的简历，匹配ATS关键词。" />
      </Helmet>

      <div className="max-w-6xl mx-auto px-4 py-10">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-900 dark:text-white mb-2">简历编辑器</h1>
          <p className="text-slate-600 dark:text-slate-400">填写信息，AI将自动优化内容</p>
        </div>

        <div className="grid lg:grid-cols-2 gap-8">
          <div className="space-y-6">
            <div className="card">
              <div className="flex items-center justify-between mb-5">
                <h2 className="text-lg font-semibold text-slate-900 dark:text-white">基本信息</h2>
                <button className="btn-secondary text-sm flex items-center gap-1.5">
                  <HiUpload className="w-4 h-4" /> 上传简历
                </button>
              </div>
              <div className="space-y-4">
                <div>
                  <label className="label">姓名</label>
                  <input className="input-field" placeholder="请输入姓名" value={resume.name} onChange={e => setResume({ ...resume, name: e.target.value })} />
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="label">邮箱</label>
                    <input type="email" className="input-field" placeholder="email@example.com" value={resume.email} onChange={e => setResume({ ...resume, email: e.target.value })} />
                  </div>
                  <div>
                    <label className="label">电话</label>
                    <input type="tel" className="input-field" placeholder="13800138000" value={resume.phone} onChange={e => setResume({ ...resume, phone: e.target.value })} />
                  </div>
                </div>
                <div>
                  <label className="label">个人简介</label>
                  <textarea className="input-field h-28 resize-none" placeholder="简要介绍你的专业背景..." value={resume.summary} onChange={e => setResume({ ...resume, summary: e.target.value })} />
                </div>
              </div>
            </div>

            <div className="card">
              <h2 className="text-lg font-semibold text-slate-900 dark:text-white mb-4">目标职位</h2>
              <div className="space-y-4">
                <div>
                  <label className="label">目标职位</label>
                  <input className="input-field" placeholder="例如：产品经理、前端工程师" value={targetJob} onChange={e => setTargetJob(e.target.value)} />
                </div>
                <div>
                  <label className="label">职位描述（可选）</label>
                  <textarea className="input-field h-28 resize-none" placeholder="粘贴职位描述..." value={jobDescription} onChange={e => setJobDescription(e.target.value)} />
                </div>
              </div>
            </div>

            <button onClick={handleOptimize} disabled={isOptimizing} className="btn-primary w-full py-3.5 text-base flex items-center justify-center gap-2">
              <HiSparkles className="w-5 h-5" />
              {isOptimizing ? 'AI优化中...' : 'AI智能优化'}
            </button>
          </div>

          <div className="space-y-6">
            <div className="card min-h-[500px]">
              <div className="flex gap-4 border-b border-slate-200 dark:border-slate-700 mb-5">
                {['edit', 'result'].map(t => (
                  <button key={t} className={`pb-3 text-sm font-medium border-b-2 transition-colors ${tab === t ? 'border-blue-500 text-blue-600' : 'border-transparent text-slate-500 hover:text-slate-700'}`} onClick={() => setTab(t)}>
                    {t === 'edit' ? '预览简历' : '优化结果'}
                  </button>
                ))}
              </div>

              {tab === 'edit' ? (
                <div className="prose prose-slate dark:prose-invert max-w-none">
                  <h3 className="text-xl font-semibold">{resume.name || '你的姓名'}</h3>
                  <p className="text-slate-500 text-sm">{resume.email} | {resume.phone}</p>
                  <p className="mt-3 text-slate-600 dark:text-slate-400">{resume.summary || '个人简介将显示在这里...'}</p>
                </div>
              ) : result ? (
                <div className="space-y-5">
                  <div className="flex items-center justify-between p-4 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800">
                    <span className="font-medium text-emerald-700 dark:text-emerald-300">ATS匹配度</span>
                    <span className="text-2xl font-bold text-emerald-600">{result.ats_score}%</span>
                  </div>
                  <div>
                    <h4 className="font-medium text-slate-900 dark:text-white mb-2">推荐关键词</h4>
                    <div className="flex flex-wrap gap-2">
                      {result.keywords?.map((kw, i) => (
                        <span key={i} className="px-3 py-1 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-300 rounded-full text-sm">{kw}</span>
                      ))}
                    </div>
                  </div>
                  <div>
                    <h4 className="font-medium text-slate-900 dark:text-white mb-2">优化建议</h4>
                    <ul className="space-y-2">
                      {result.suggestions?.map((s, i) => (
                        <li key={i} className="text-sm text-slate-600 dark:text-slate-400 flex items-start gap-2">
                          <span className="text-blue-500 mt-1">•</span>{s}
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              ) : (
                <div className="flex flex-col items-center justify-center h-64 text-slate-400">
                  <HiSparkles className="w-10 h-10 mb-3 opacity-50" />
                  <p className="text-sm">点击"AI智能优化"查看结果</p>
                </div>
              )}
            </div>

            <button className="btn-primary w-full py-3.5 text-base flex items-center justify-center gap-2">
              <HiDocumentArrowDown className="w-5 h-5" /> 导出PDF
            </button>
          </div>
        </div>
      </div>
    </>
  )
}

export default EditorPage
