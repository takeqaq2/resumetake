import React from 'react'
import { Link } from 'react-router-dom'

function Footer() {
  return (
    <footer className="bg-slate-900 text-slate-400">
      <div className="max-w-6xl mx-auto px-4 py-12">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-10">
          <div className="col-span-2 md:col-span-1">
            <Link to="/" className="flex items-center gap-2 mb-4">
              <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-violet-500 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-sm">R</span>
              </div>
              <span className="text-white font-semibold">ResumeTake</span>
            </Link>
            <p className="text-sm leading-relaxed">AI智能简历优化工具，助你获得理想工作。</p>
          </div>
          <div>
            <h4 className="text-white font-medium mb-3 text-sm">产品</h4>
            <ul className="space-y-2 text-sm">
              <li><Link to="/editor" className="hover:text-white transition-colors">AI简历优化</Link></li>
              <li><Link to="/templates" className="hover:text-white transition-colors">简历模板</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-white font-medium mb-3 text-sm">支持</h4>
            <ul className="space-y-2 text-sm">
              <li><span className="cursor-default">使用教程</span></li>
              <li><span className="cursor-default">常见问题</span></li>
              <li><span className="cursor-default">联系我们</span></li>
            </ul>
          </div>
          <div>
            <h4 className="text-white font-medium mb-3 text-sm">法律</h4>
            <ul className="space-y-2 text-sm">
              <li><span className="cursor-default">隐私政策</span></li>
              <li><span className="cursor-default">服务条款</span></li>
            </ul>
          </div>
        </div>
        <div className="border-t border-slate-800 pt-6 text-center text-sm">
          <p>&copy; {new Date().getFullYear()} ResumeTake. All rights reserved.</p>
        </div>
      </div>
    </footer>
  )
}

export default Footer
