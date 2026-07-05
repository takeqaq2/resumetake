# ResumeTake - AI简历优化工具

> 智能AI简历优化，一键生成专业简历，ATS关键词匹配，提升求职成功率。

## 功能特点

- **AI智能优化** - 基于先进AI技术，自动优化简历内容和格式
- **ATS关键词匹配** - 智能分析目标职位，匹配ATS关键词
- **专业模板** - 多种模板覆盖各行各业
- **PDF导出** - 一键导出高质量PDF简历
- **暗色模式** - 支持亮色/暗色主题切换

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | React 18 + TailwindCSS + Vite |
| 后端 | Python FastAPI (async) |
| 部署 | Docker + Nginx |
| SEO | Schema.org + Open Graph + Sitemap |

## 快速开始

```bash
# 克隆项目
git clone https://github.com/yourusername/resumetake.git
cd resumetake

# Docker启动
docker-compose up -d --build

# 访问
open http://localhost:8088
```

## 项目结构

```
resumetake/
├── frontend/          # React前端
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   └── App.jsx
│   └── public/
│       ├── robots.txt
│       └── sitemap.xml
├── backend/           # FastAPI后端
│   └── app/
│       └── main.py
├── nginx/             # Nginx配置
└── docker-compose.yml
```

## SEO优化

- Schema.org结构化数据 (WebApplication + FAQPage + Organization)
- Open Graph社交分享标签
- 语义化HTML + 响应式设计
- 自动生成sitemap.xml和robots.txt
- 页面加载性能优化

## License

MIT
