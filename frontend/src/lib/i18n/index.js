export const LANGUAGES = {
  zh: { name: '中文', dir: 'ltr', flag: '🇨🇳', country: 'China', locale: 'zh_CN' },
  en: { name: 'English', dir: 'ltr', flag: '🇺🇸', country: 'United States', locale: 'en_US' },
  ja: { name: '日本語', dir: 'ltr', flag: '🇯🇵', country: 'Japan', locale: 'ja_JP' },
  ko: { name: '한국어', dir: 'ltr', flag: '🇰🇷', country: 'South Korea', locale: 'ko_KR' },
  es: { name: 'Español', dir: 'ltr', flag: '🇪🇸', country: 'Spain', locale: 'es_ES' },
  pt: { name: 'Português', dir: 'ltr', flag: '🇧🇷', country: 'Brazil', locale: 'pt_BR' },
  fr: { name: 'Français', dir: 'ltr', flag: '🇫🇷', country: 'France', locale: 'fr_FR' },
  de: { name: 'Deutsch', dir: 'ltr', flag: '🇩🇪', country: 'Germany', locale: 'de_DE' },
  ar: { name: 'العربية', dir: 'rtl', flag: '🇸🇦', country: 'Saudi Arabia', locale: 'ar_SA' },
  hi: { name: 'हिन्दी', dir: 'ltr', flag: '🇮🇳', country: 'India', locale: 'hi_IN' }
};

export const translations = {
  zh: {
    meta: {
      title: 'AI简历优化工具 - ResumeTake | 智能简历制作平台',
      description: 'AI智能简历优化工具，一键生成专业简历，ATS关键词匹配，提升求职成功率。免费在线简历制作，支持PDF导出。',
      keywords: 'AI简历优化,简历生成器,智能简历制作,ATS简历优化,免费简历工具,在线简历编辑,一键生成简历,求职简历',
      editorTitle: '简历编辑器 - ResumeTake | AI简历优化',
      editorDesc: '使用AI智能优化你的简历，匹配ATS关键词，提升求职成功率。',
      editorKeywords: '简历编辑器,AI简历优化,在线简历编辑,ATS优化',
      templatesTitle: '简历模板 - ResumeTake | 免费专业简历模板',
      templatesDesc: '多种专业简历模板，覆盖互联网、金融、教育等行业，免费使用。',
      templatesKeywords: '简历模板,免费简历模板,专业简历模板,简历设计'
    },
    nav: {
      home: '首页',
      start: '免费开始',
      optimize: 'AI优化',
      generate: '0基础生成',
      jobs: '校招信息',
      pricing: '付费方案',
      createResume: '创建简历',
      templates: '模板',
      language: '语言',
      login: '登录',
      logout: '退出'
    },
    hero: {
      badge: 'AI驱动的智能简历优化',
      title1: '让AI帮你打造',
      title2: '完美简历',
      subtitle: '上传简历，AI自动优化内容、匹配ATS关键词、提升通过率。一键生成专业简历。',
      cta: '立即开始 →',
      learnMore: '了解更多'
    },
    features: {
      title: '为什么选择 ResumeTake',
      subtitle: '基于先进AI技术，让简历优化变得简单高效',
      items: [
        { icon: '⚡', title: 'AI智能优化', desc: '自动优化简历内容、用词和格式，突出核心竞争力与成就' },
        { icon: '🔍', title: 'ATS关键词匹配', desc: '智能分析目标职位，自动匹配ATS关键词，提升通过率' },
        { icon: '📄', title: '专业模板', desc: '多种模板覆盖各行各业，一键切换风格，导出PDF' }
      ]
    },
    cta: {
      title: '开始优化你的简历',
      subtitle: '免费使用AI简历优化工具，提升求职成功率',
      button: '免费开始 →'
    },
    editor: {
      title: 'AI简历优化',
      subtitle: '粘贴简历内容，选择优化方向，一键获得专业建议',
      pasteResume: '粘贴简历内容',
      pasteResumePlaceholder: '将你的简历内容粘贴到这里...\n\n支持纯文本格式，包括个人信息、工作经历、教育背景、技能等。',
      pasteResumeHint: '直接粘贴简历全文即可',
      targetJob: '目标职位',
      targetJobPlaceholder: '例如：产品经理、前端工程师',
      jobDesc: '职位描述（可选）',
      jobDescPlaceholder: '粘贴职位描述，AI将针对性优化...',
      optModules: '优化模块',
      optModulesHint: '选择你需要的优化方向',
      module_ats: 'ATS关键词匹配',
      module_ats_desc: '提取并匹配目标职位的ATS关键词',
      module_star: 'STAR法则优化',
      module_star_desc: '用STAR法则重写工作经历',
      module_quant: '量化成果',
      module_quant_desc: '添加数据和量化指标突出成就',
      module_summary: '个人简介优化',
      module_summary_desc: '优化自我介绍，突出核心竞争力',
      module_format: '格式与排版',
      module_format_desc: '优化简历结构和排版格式',
      selectAll: '全选',
      deselectAll: '取消全选',
      optimizing: 'AI优化中...',
      optimizeBtn: '✨ 一键AI优化',
      previewTab: '原始简历',
      resultTab: '优化结果',
      defaultName: '你的姓名',
      defaultSummary: '个人简介将显示在这里...',
      atsScore: 'ATS匹配度',
      keywords: '推荐关键词',
      suggestions: '优化建议',
      exportPdf: '📥 导出PDF',
      emptyResult: '粘贴简历并点击优化查看结果',
      pasteFirst: '请先粘贴简历内容',
      optimized: '优化完成！',
      optimizedTime: '用时',
      orUploadResume: '或上传简历文件',
      uploadFile: '上传文件',
      uploadHint: '支持 TXT、PDF、DOC、DOCX 格式，最大 5MB',
      dragDrop: '拖拽文件到此处，或点击上传',
      uploading: '上传中...',
      uploadSuccess: '文件上传成功',
      uploadError: '上传失败，请重试',
      fetchJobUrl: '从URL获取职位信息',
      jobUrlPlaceholder: '粘贴职位页面URL，如 https://...',
      fetching: '获取中...',
      fetchSuccess: '职位内容已获取',
      fetchError: '获取失败，请检查URL',
      orPasteUrl: '或粘贴职位链接'
    },
    templates: {
      title: '专业简历模板',
      subtitle: '选择适合你行业的模板，一键套用',
      items: [
        { id: 'professional', name: '专业商务', desc: '适合传统行业和商务岗位' },
        { id: 'modern', name: '现代简约', desc: '适合互联网和科技行业' },
        { id: 'creative', name: '创意设计', desc: '适合设计和创意岗位' },
        { id: 'academic', name: '学术科研', desc: '适合教育和研究岗位' },
        { id: 'executive', name: '高管专用', desc: '适合高级管理岗位' },
        { id: 'minimal', name: '极简风格', desc: '简洁大方，通用性强' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: '创建简历',
      templates: '模板'
    },
    auth: {
      login: '登录',
      register: '注册',
      email: '邮箱',
      password: '密码',
      name: '姓名',
      loginBtn: '登录',
      registerBtn: '注册',
      noAccount: '没有账户？',
      hasAccount: '已有账户？'
    },
    jobs: {
      title: '职位中心',
      subtitle: '发现最新的工作机会',
      search: '搜索职位、公司或地点...',
      apply: '申请',
      fullTime: '全职',
      intern: '实习'
    },
    generate: {
      title: '0基础生成简历',
      subtitle: 'AI引导你一步步构建专业简历',
      placeholder: '输入你的信息...',
      send: '发送'
    },
    perspective: {
      title: '4视角分析',
      original: '原始的我',
      optimized: '优化后的我',
      imagined: '我幻想的我',
      desired: 'HR希望的我'
    },
    pricing: {
      title: '定价方案',
      subtitle: '选择最适合你的方案',
      free: '免费版',
      pro: '专业版',
      enterprise: '企业版',
      monthly: '月',
      current: '当前方案',
      upgrade: '升级',
      popular: '最受欢迎'
    },
    faq: [
      { q: 'ResumeTake是什么？', a: 'ResumeTake是一款AI智能简历优化工具，帮助用户一键生成专业简历，自动匹配ATS关键词，提升求职成功率。' },
      { q: 'ResumeTake是免费的吗？', a: '是的，ResumeTake提供免费的基础AI简历优化功能。' },
      { q: '如何使用AI优化简历？', a: '上传简历或填写信息，输入目标职位，点击AI优化按钮即可获得优化建议。' }
    ]
  },
  en: {
    meta: {
      title: 'AI Resume Builder - ResumeTake | Smart Resume Optimization',
      description: 'AI-powered resume optimizer. Generate professional resumes, match ATS keywords, and boost your job search success. Free online resume builder with PDF export.',
      keywords: 'AI resume builder,resume optimizer,ATS resume,free resume tool,online resume editor,professional resume,job search',
      editorTitle: 'Resume Editor - ResumeTake | AI Resume Optimization',
      editorDesc: 'Optimize your resume with AI, match ATS keywords, and boost your job search success.',
      editorKeywords: 'resume editor,AI resume optimization,online resume editor,ATS optimization',
      templatesTitle: 'Resume Templates - ResumeTake | Free Professional Templates',
      templatesDesc: 'Professional resume templates for every industry. Free to use.',
      templatesKeywords: 'resume templates,free resume templates,professional resume,resume design'
    },
    nav: {
      home: 'Home',
      start: 'Get Started',
      optimize: 'AI Optimize',
      generate: 'Generate',
      jobs: 'Jobs',
      pricing: 'Pricing',
      createResume: 'Create Resume',
      templates: 'Templates',
      language: 'Language',
      login: 'Login',
      logout: 'Logout'
    },
    hero: {
      badge: 'AI-Powered Smart Resume Optimization',
      title1: 'Let AI Build Your',
      title2: 'Perfect Resume',
      subtitle: 'Upload your resume, and AI automatically optimizes content, matches ATS keywords, and improves your pass rate. Generate a professional resume in one click.',
      cta: 'Get Started →',
      learnMore: 'Learn More'
    },
    features: {
      title: 'Why Choose ResumeTake',
      subtitle: 'Powered by advanced AI technology to make resume optimization simple and effective',
      items: [
        { icon: '⚡', title: 'AI Smart Optimization', desc: 'Automatically optimize resume content, wording, and format to highlight core competencies' },
        { icon: '🔍', title: 'ATS Keyword Matching', desc: 'Intelligently analyze target jobs and auto-match ATS keywords to boost pass rates' },
        { icon: '📄', title: 'Professional Templates', desc: 'Multiple templates for every industry, one-click style switching, PDF export' }
      ]
    },
    cta: {
      title: 'Start Optimizing Your Resume',
      subtitle: 'Use our free AI resume optimizer to boost your job search success',
      button: 'Get Started Free →'
    },
    editor: {
      title: 'AI Resume Optimizer',
      subtitle: 'Paste your resume, choose optimization modules, get professional results',
      pasteResume: 'Paste Resume Content',
      pasteResumePlaceholder: 'Paste your resume content here...\n\nSupports plain text: personal info, work experience, education, skills, etc.',
      pasteResumeHint: 'Paste your full resume text',
      targetJob: 'Target Position',
      targetJobPlaceholder: 'e.g., Product Manager, Frontend Engineer',
      jobDesc: 'Job Description (Optional)',
      jobDescPlaceholder: 'Paste the job description for targeted optimization...',
      optModules: 'Optimization Modules',
      optModulesHint: 'Choose what to optimize',
      module_ats: 'ATS Keyword Matching',
      module_ats_desc: 'Extract and match ATS keywords for the target job',
      module_star: 'STAR Method Rewrite',
      module_star_desc: 'Rewrite work experience using STAR method',
      module_quant: 'Quantify Achievements',
      module_quant_desc: 'Add data and metrics to highlight achievements',
      module_summary: 'Summary Optimization',
      module_summary_desc: 'Optimize professional summary and core competencies',
      module_format: 'Format & Layout',
      module_format_desc: 'Optimize resume structure and formatting',
      selectAll: 'Select All',
      deselectAll: 'Deselect All',
      optimizing: 'AI Optimizing...',
      optimizeBtn: '✨ AI Optimize Now',
      previewTab: 'Original Resume',
      resultTab: 'Optimized Result',
      defaultName: 'Your Name',
      defaultSummary: 'Your summary will appear here...',
      atsScore: 'ATS Match Score',
      keywords: 'Recommended Keywords',
      suggestions: 'Optimization Suggestions',
      exportPdf: '📥 Export PDF',
      emptyResult: 'Paste resume and click Optimize to see results',
      pasteFirst: 'Please paste resume content first',
      optimized: 'Optimization Complete!',
      optimizedTime: 'Time taken',
      orUploadResume: 'Or upload a resume file',
      uploadFile: 'Upload File',
      uploadHint: 'Supports TXT, PDF, DOC, DOCX, max 5MB',
      dragDrop: 'Drag & drop or click to upload',
      uploading: 'Uploading...',
      uploadSuccess: 'File uploaded successfully',
      uploadError: 'Upload failed, please try again',
      fetchJobUrl: 'Fetch from URL',
      jobUrlPlaceholder: 'Paste job posting URL, e.g. https://...',
      fetching: 'Fetching...',
      fetchSuccess: 'Job content fetched',
      fetchError: 'Failed to fetch, please check URL',
      orPasteUrl: 'Or paste a job posting link'
    },
    templates: {
      title: 'Professional Resume Templates',
      subtitle: 'Choose a template that fits your industry',
      items: [
        { id: 'professional', name: 'Professional', desc: 'For traditional and business roles' },
        { id: 'modern', name: 'Modern', desc: 'For tech and startup roles' },
        { id: 'creative', name: 'Creative', desc: 'For design and creative roles' },
        { id: 'academic', name: 'Academic', desc: 'For education and research roles' },
        { id: 'executive', name: 'Executive', desc: 'For senior management roles' },
        { id: 'minimal', name: 'Minimal', desc: 'Clean and versatile' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: 'Create Resume',
      templates: 'Templates'
    },
    auth: {
      login: 'Login',
      register: 'Register',
      email: 'Email',
      password: 'Password',
      name: 'Name',
      loginBtn: 'Login',
      registerBtn: 'Register',
      noAccount: "Don't have an account?",
      hasAccount: 'Already have an account?'
    },
    jobs: {
      title: 'Job Board',
      subtitle: 'Discover the latest job opportunities',
      search: 'Search jobs, companies, or locations...',
      apply: 'Apply',
      fullTime: 'Full-time',
      intern: 'Intern'
    },
    generate: {
      title: '0-Basis Resume Generator',
      subtitle: 'AI guides you step by step to build a professional resume',
      placeholder: 'Type your information...',
      send: 'Send'
    },
    perspective: {
      title: '4-Perspective Analysis',
      original: 'Original Me',
      optimized: 'Optimized Me',
      imagined: 'Imagined Me',
      desired: 'HR Desired Me'
    },
    pricing: {
      title: 'Pricing',
      subtitle: 'Choose the plan that fits you best',
      free: 'Free',
      pro: 'Pro',
      enterprise: 'Enterprise',
      monthly: 'mo',
      current: 'Current Plan',
      upgrade: 'Upgrade',
      popular: 'Most Popular'
    },
    faq: [
      { q: 'What is ResumeTake?', a: 'ResumeTake is an AI-powered resume optimization tool that helps users generate professional resumes, automatically match ATS keywords, and improve job search success rates.' },
      { q: 'Is ResumeTake free?', a: 'Yes, ResumeTake offers free basic AI resume optimization features.' },
      { q: 'How to use AI resume optimization?', a: 'Upload your resume or fill in your details, enter your target position, and click the AI optimize button to get optimization suggestions.' }
    ]
  },
  ja: {
    meta: {
      title: 'AI履歴書作成ツール - ResumeTake | スマート履歴書最適化',
      description: 'AI搭載の履歴書最適化ツール。プロフェッショナルな履歴書をワンクリックで作成、ATSキーワードマッチングで就職成功率を向上。無料でPDFエクスポート対応。',
      keywords: 'AI履歴書,履歴書作成,ATS履歴書,無料履歴書ツール,オンライン履歴書編集,プロ履歴書,就職活動',
      editorTitle: '履歴書エディタ - ResumeTake | AI履歴書最適化',
      editorDesc: 'AIで履歴書を最適化、ATSキーワードをマッチングし、就職成功率を向上させます。',
      editorKeywords: '履歴書エディタ,AI履歴書最適化,オンライン履歴書編集,ATS最適化',
      templatesTitle: '履歴書テンプレート - ResumeTake | 無料プロテンプレート',
      templatesDesc: 'あらゆる業界に対応するプロフェッショナルな履歴書テンプレート。無料でご利用いただけます。',
      templatesKeywords: '履歴書テンプレート,無料テンプレート,プロテンプレート,履歴書デザイン'
    },
    nav: {
      home: 'ホーム',
      start: '無料で始める',
      optimize: 'AI最適化',
      generate: 'ゼロから生成',
      jobs: '求人情報',
      pricing: '料金プラン',
      createResume: '履歴書作成',
      templates: 'テンプレート',
      language: '言語',
      login: 'ログイン',
      logout: 'ログアウト'
    },
    hero: {
      badge: 'AI搭載のスマート履歴書最適化',
      title1: 'AIがあなたの',
      title2: '完美な履歴書を構築',
      subtitle: '履歴書をアップロードするだけで、AIが自動でコンテンツを最適化し、ATSキーワードをマッチングし、通過率を向上させます。ワンクリックでプロの履歴書を作成。',
      cta: '無料で始める →',
      learnMore: '詳しく見る'
    },
    features: {
      title: 'ResumeTakeが選ばれる理由',
      subtitle: '先進AI技術で履歴書の最適化をシンプルで効果的に',
      items: [
        { icon: '⚡', title: 'AIスマート最適化', desc: '履歴書の内容、表現、フォーマットを自動最適化し、核心的な竞争力を強調' },
        { icon: '🔍', title: 'ATSキーワードマッチング', desc: '求人をインテリジェントに分析し、ATSキーワードを自動マッチング' },
        { icon: '📄', title: 'プロテンプレート', desc: 'あらゆる業界に対応するテンプレート、ワンクリックでスタイル切替、PDF出力' }
      ]
    },
    cta: {
      title: '履歴書の最適化を開始',
      subtitle: '無料のAI履歴書最適化ツールで就職成功率を向上',
      button: '無料で始める →'
    },
    editor: {
      title: 'AI履歴書最適化',
      subtitle: '履歴書を貼り付け、最適化モジュールを選んでプロの結果を取得',
      pasteResume: '履歴書を貼り付け',
      pasteResumePlaceholder: '履歴書の内容をここに貼り付けてください...\n\nテキスト形式対応：個人情報、職務経歴、学歴、スキルなど',
      pasteResumeHint: '履歴書全文を貼り付けてください',
      targetJob: '希望職種',
      targetJobPlaceholder: '例：プロダクトマネージャー、フロントエンドエンジニア',
      jobDesc: '職務記述書（オプション）',
      jobDescPlaceholder: '職務記述書を貼り付けると、より的確な最適化が可能になります...',
      optModules: '最適化モジュール',
      optModulesHint: '最適化したい項目を選択',
      module_ats: 'ATSキーワードマッチング',
      module_ats_desc: '希望職種のATSキーワードを抽出・マッチング',
      module_star: 'STAR法による最適化',
      module_star_desc: 'STAR法で職務経歴を書き直し',
      module_quant: '成果の数値化',
      module_quant_desc: 'データや指標を追加して成果を強調',
      module_summary: '自己PR最適化',
      module_summary_desc: '自己PRと核心竞争力を最適化',
      module_format: 'フォーマット',
      module_format_desc: '履歴書の構成とフォーマットを最適化',
      selectAll: 'すべて選択',
      deselectAll: 'すべて解除',
      optimizing: 'AI最適化中...',
      optimizeBtn: '✨ 今すぐ最適化',
      previewTab: '元の履歴書',
      resultTab: '最適化結果',
      defaultName: 'お名前',
      defaultSummary: 'ここに自己PRが表示されます...',
      atsScore: 'ATSマッチングスコア',
      keywords: '推奨キーワード',
      suggestions: '最適化提案',
      exportPdf: '📥 PDFエクスポート',
      emptyResult: '履歴書を貼り付けて最適化をクリック',
      pasteFirst: 'まず履歴書の内容を貼り付けてください',
      optimized: '最適化完了！',
      optimizedTime: '所要時間',
      orUploadResume: 'または履歴書ファイルをアップロード',
      uploadFile: 'ファイルアップロード',
      uploadHint: 'TXT、PDF、DOC、DOCX対応、最大5MB',
      dragDrop: 'ドラッグ&ドロップまたはクリックしてアップロード',
      uploading: 'アップロード中...',
      uploadSuccess: 'アップロード成功',
      uploadError: 'アップロード失敗、再試行してください',
      fetchJobUrl: 'URLから取得',
      jobUrlPlaceholder: '求人ページのURLを貼り付け、例：https://...',
      fetching: '取得中...',
      fetchSuccess: '求人内容を取得しました',
      fetchError: '取得失敗、URLを確認してください',
      orPasteUrl: 'または求人リンクを貼り付け'
    },
    templates: {
      title: 'プロフェッショナルな履歴書テンプレート',
      subtitle: 'あなたの業界に合ったテンプレートを選択',
      items: [
        { id: 'professional', name: 'プロフェッショナル', desc: '伝統的・ビジネス職向け' },
        { id: 'modern', name: 'モダン', desc: 'IT・テック業界向け' },
        { id: 'creative', name: 'クリエイティブ', desc: 'デザイン・クリエイティブ職向け' },
        { id: 'academic', name: 'アカデミック', desc: '教育・研究職向け' },
        { id: 'executive', name: 'エグゼクティブ', desc: '上級管理職向け' },
        { id: 'minimal', name: 'ミニマル', desc: 'シンプルで汎用性が高い' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: '履歴書作成',
      templates: 'テンプレート'
    },
    auth: {
      login: 'ログイン',
      register: '新規登録',
      email: 'メールアドレス',
      password: 'パスワード',
      name: '氏名',
      loginBtn: 'ログイン',
      registerBtn: '新規登録',
      noAccount: 'アカウントをお持ちでないですか？',
      hasAccount: 'アカウントをお持ちですか？'
    },
    jobs: {
      title: '求人情報',
      subtitle: '最新の求人情報を見つけましょう',
      search: '求人、会社名、場所で検索...',
      apply: '応募',
      fullTime: '正社員',
      intern: 'インターン'
    },
    generate: {
      title: 'ゼロから履歴書作成',
      subtitle: 'AIがステップバイステップでプロの履歴書作成をガイド',
      placeholder: '情報を入力してください...',
      send: '送信'
    },
    perspective: {
      title: '4視点分析',
      original: '今の私',
      optimized: '最適化された私',
      imagined: '理想の私',
      desired: '企業が求める私'
    },
    pricing: {
      title: '料金プラン',
      subtitle: 'あなたに最適なプランを選択',
      free: '無料版',
      pro: 'プロ版',
      enterprise: 'エンタープライズ',
      monthly: '月',
      current: '現在のプラン',
      upgrade: 'アップグレード',
      popular: '一番人気'
    },
    faq: [
      { q: 'ResumeTakeとは？', a: 'ResumeTakeは、プロフェッショナルな履歴書をワンクリックで作成し、ATSキーワードを自動マッチングして就職成功率を向上させるAI搭載の履歴書最適化ツールです。' },
      { q: 'ResumeTakeは無料ですか？', a: 'はい、ResumeTakeは基本的なAI履歴書最適化機能を無料で提供しています。' },
      { q: 'AI履歴書最適化の使い方は？', a: '履歴書をアップロードするか情報を入力し、希望職種を入力してAI最適化ボタンをクリックすると、最適化提案が得られます。' }
    ]
  },
  ko: {
    meta: {
      title: 'AI 이력서 작성 도구 - ResumeTake | 스마트 이력서 최적화',
      description: 'AI 기반 이력서 최적화 도구. 전문 이력서를 원클릭으로 생성하고, ATS 키워드 매칭으로 취업 성공률을 높이세요. 무료 PDF 내보내기 지원.',
      keywords: 'AI 이력서,이력서 작성,ATS 이력서,무료 이력서 도구,온라인 이력서 편집,전문 이력서,취업',
      editorTitle: '이력서 편집기 - ResumeTake | AI 이력서 최적화',
      editorDesc: 'AI로 이력서를 최적화하고, ATS 키워드를 매칭하여 취업 성공률을 높이세요.',
      editorKeywords: '이력서 편집기,AI 이력서 최적화,온라인 이력서 편집,ATS 최적화',
      templatesTitle: '이력서 템플릿 - ResumeTake | 무료 전문 템플릿',
      templatesDesc: '모든 산업에 맞는 전문 이력서 템플릿. 무료로 사용하세요.',
      templatesKeywords: '이력서 템플릿,무료 이력서 템플릿,전문 이력서,이력서 디자인'
    },
    nav: {
      home: '홈',
      start: '무료로 시작',
      optimize: 'AI 최적화',
      generate: '이력서 생성',
      jobs: '취업 정보',
      pricing: '요금제',
      createResume: '이력서 만들기',
      templates: '템플릿',
      language: '언어',
      login: '로그인',
      logout: '로그아웃'
    },
    hero: {
      badge: 'AI 기반 스마트 이력서 최적화',
      title1: 'AI가 만들어주는',
      title2: '완벽한 이력서',
      subtitle: '이력서를 업로드하면 AI가 자동으로 내용을 최적화하고, ATS 키워드를 매칭하며, 합격률을 높여줍니다. 원클릭으로 전문 이력서를 만들어보세요.',
      cta: '무료로 시작 →',
      learnMore: '자세히 보기'
    },
    features: {
      title: 'ResumeTake가 선택되는 이유',
      subtitle: ' tiên진 AI 기술로 이력서 최적화를 간단하고 효과적으로',
      items: [
        { icon: '⚡', title: 'AI 스마트 최적화', desc: '이력서 내용, 표현, 포맷을 자동 최적화하여 핵심 역량을 강조' },
        { icon: '🔍', title: 'ATS 키워드 매칭', desc: '공고를 지능적으로 분석하고 ATS 키워드를 자동 매칭' },
        { icon: '📄', title: '전문 템플릿', desc: '모든 산업에 맞는 템플릿, 원클릭 스타일 전환, PDF 내보내기' }
      ]
    },
    cta: {
      title: '이력서 최적화를 시작하세요',
      subtitle: '무료 AI 이력서 최적화 도구로 취업 성공률을 높이세요',
      button: '무료로 시작 →'
    },
    editor: {
      title: 'AI 이력서 최적화',
      subtitle: '이력서를 붙여넣고 최적화 모듈을 선택하여 전문 결과를 받으세요',
      pasteResume: '이력서 내용 붙여넣기',
      pasteResumePlaceholder: '이력서 내용을 여기에 붙여넣으세요...\n\n텍스트 형식 지원: 개인정보, 업무 경험, 학력, 기술 등',
      pasteResumeHint: '이력서 전체를 붙여넣으세요',
      targetJob: '희망 직종',
      targetJobPlaceholder: '예: 프로덕트 매니저, 프론트엔드 엔지니어',
      jobDesc: '직무 설명 (선택사항)',
      jobDescPlaceholder: '직무 설명을 붙여넣으면 더 정확한 최적화가 가능합니다...',
      optModules: '최적화 모듈',
      optModulesHint: '최적화할 항목을 선택하세요',
      module_ats: 'ATS 키워드 매칭',
      module_ats_desc: '희망 직종의 ATS 키워드를 추출하고 매칭',
      module_star: 'STAR 방법 최적화',
      module_star_desc: 'STAR 방법으로 업무 경험을 재작성',
      module_quant: '성과 수량화',
      module_quant_desc: '데이터와 지표를 추가하여 성과 강조',
      module_summary: '자기소개 최적화',
      module_summary_desc: '자기소개와 핵심 역량을 최적화',
      module_format: '포맷 및 레이아웃',
      module_format_desc: '이력서 구조와 포맷을 최적화',
      selectAll: '전체 선택',
      deselectAll: '전체 해제',
      optimizing: 'AI 최적화 중...',
      optimizeBtn: '✨ 지금 최적화하기',
      previewTab: '원본 이력서',
      resultTab: '최적화 결과',
      defaultName: '이름',
      defaultSummary: '여기에 자기소개가 표시됩니다...',
      atsScore: 'ATS 매칭 점수',
      keywords: '추천 키워드',
      suggestions: '최적화 제안',
      exportPdf: '📥 PDF 내보내기',
      emptyResult: '이력서를 붙여넣고 최적화를 클릭하세요',
      pasteFirst: '이력서 내용을 먼저 붙여넣으세요',
      optimized: '최적화 완료!',
      optimizedTime: '소요 시간',
      orUploadResume: '또는 이력서 파일 업로드',
      uploadFile: '파일 업로드',
      uploadHint: 'TXT, PDF, DOC, DOCX 지원, 최대 5MB',
      dragDrop: '드래그 앤 드롭 또는 클릭하여 업로드',
      uploading: '업로드 중...',
      uploadSuccess: '업로드 성공',
      uploadError: '업로드 실패, 다시 시도해주세요',
      fetchJobUrl: 'URL에서 가져오기',
      jobUrlPlaceholder: '채용 공고 URL을 붙여넣으세요, 예: https://...',
      fetching: '가져오는 중...',
      fetchSuccess: '채용 내용을 가져왔습니다',
      fetchError: '가져오기 실패, URL을 확인하세요',
      orPasteUrl: '또는 채용 링크를 붙여넣으세요'
    },
    templates: {
      title: '전문 이력서 템플릿',
      subtitle: '귀하의 산업에 맞는 템플릿을 선택하세요',
      items: [
        { id: 'professional', name: '프로페셔널', desc: '전통적 및 비즈니스 직무용' },
        { id: 'modern', name: '모던', desc: 'IT 및 테크 업계용' },
        { id: 'creative', name: '크리에이티브', desc: '디자인 및 크리에이티브 직무용' },
        { id: 'academic', name: '아카데믹', desc: '교육 및 연구 직무용' },
        { id: 'executive', name: '임원급', desc: '고위 경영진용' },
        { id: 'minimal', name: '미니멀', desc: '깔끔하고 범용적' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: '이력서 만들기',
      templates: '템플릿'
    },
    auth: {
      login: '로그인',
      register: '회원가입',
      email: '이메일',
      password: '비밀번호',
      name: '이름',
      loginBtn: '로그인',
      registerBtn: '회원가입',
      noAccount: '계정이 없으신가요?',
      hasAccount: '이미 계정이 있으신가요?'
    },
    jobs: {
      title: '채용 게시판',
      subtitle: '최신 채용 기회를 찾아보세요',
      search: '직무, 회사, 지역으로 검색...',
      apply: '지원',
      fullTime: '정규직',
      intern: '인턴'
    },
    generate: {
      title: '0부터 이력서 생성',
      subtitle: 'AI가 단계별로 전문 이력서 작성을 안내합니다',
      placeholder: '정보를 입력하세요...',
      send: '보내기'
    },
    perspective: {
      title: '4관점 분석',
      original: '원래의 나',
      optimized: '최적화된 나',
      imagined: '상상의 나',
      desired: '기업이 원하는 나'
    },
    pricing: {
      title: '요금제',
      subtitle: '가장 적합한 플랜을 선택하세요',
      free: '무료版',
      pro: '프로版',
      enterprise: '엔터프라이즈',
      monthly: '월',
      current: '현재 플랜',
      upgrade: '업그레이드',
      popular: '가장 인기'
    },
    faq: [
      { q: 'ResumeTake란?', a: 'ResumeTake는 AI 기반 이력서 최적화 도구로, 전문 이력서를 원클릭으로 생성하고 ATS 키워드를 자동 매칭하여 취업 성공률을 높여줍니다.' },
      { q: 'ResumeTake는 무료인가요?', a: '네, ResumeTake는 기본 AI 이력서 최적화 기능을 무료로 제공합니다.' },
      { q: 'AI 이력서 최적화 사용법은?', a: '이력서를 업로드하거나 정보를 입력하고, 희망 직종을 입력한 후 AI 최적화 버튼을 클릭하면 최적화 제안을 받을 수 있습니다.' }
    ]
  },
  es: {
    meta: {
      title: 'Constructor de CV con IA - ResumeTake | Optimización Inteligente de CV',
      description: 'Optimizador de CV con IA. Genera CVs profesionales, coincide con palabras clave ATS y aumenta tu éxito en la búsqueda de empleo. Gratis con exportación a PDF.',
      keywords: 'CV con IA,optimizador de CV,CV ATS,creador de CV gratis,editor de CV online,CV profesional,búsqueda de empleo',
      editorTitle: 'Editor de CV - ResumeTake | Optimización IA de CV',
      editorDesc: 'Optimiza tu CV con IA, coincide con palabras clave ATS y aumenta tu éxito laboral.',
      editorKeywords: 'editor de CV,optimización IA de CV,editor online de CV,optimización ATS',
      templatesTitle: 'Plantillas de CV - ResumeTake | Plantillas Profesionales Gratis',
      templatesDesc: 'Plantillas de CV profesionales para cada sector. Gratis.',
      templatesKeywords: 'plantillas CV,plantillas CV gratis,CV profesional,diseño CV'
    },
    nav: {
      home: 'Inicio',
      start: 'Comenzar',
      optimize: 'Optimizar IA',
      generate: 'Generar',
      jobs: 'Empleos',
      pricing: 'Precios',
      createResume: 'Crear CV',
      templates: 'Plantillas',
      language: 'Idioma',
      login: 'Iniciar sesión',
      logout: 'Cerrar sesión'
    },
    hero: {
      badge: 'Optimización Inteligente de CV con IA',
      title1: 'Deja que la IA construya tu',
      title2: 'CV Perfecto',
      subtitle: 'Sube tu CV y la IA optimiza automáticamente el contenido, coincide con palabras clave ATS y mejora tu tasa de aprobación. Genera un CV profesional con un clic.',
      cta: 'Comenzar →',
      learnMore: 'Saber más'
    },
    features: {
      title: 'Por qué elegir ResumeTake',
      subtitle: 'Impulsado por tecnología IA avanzada para hacer la optimización de CV simple y efectiva',
      items: [
        { icon: '⚡', title: 'Optimización IA Inteligente', desc: 'Optimiza automáticamente el contenido, redacción y formato del CV para destacar competencias clave' },
        { icon: '🔍', title: 'Coincidencia de Palabras Clave ATS', desc: 'Analiza inteligentemente las ofertas y coincide automáticamente con palabras clave ATS' },
        { icon: '📄', title: 'Plantillas Profesionales', desc: 'Múltiples plantillas para cada sector, cambio de estilo con un clic, exportación PDF' }
      ]
    },
    cta: {
      title: 'Comienza a optimizar tu CV',
      subtitle: 'Usa nuestro optimizador de CV con IA gratis para aumentar tu éxito laboral',
      button: 'Comenzar Gratis →'
    },
    editor: {
      title: 'Optimización IA de CV',
      subtitle: 'Pega tu CV, elige módulos de optimización y obtén resultados profesionales',
      pasteResume: 'Pegar Contenido del CV',
      pasteResumePlaceholder: 'Pega el contenido de tu CV aquí...\n\nSoporta texto plano: información personal, experiencia, educación, habilidades, etc.',
      pasteResumeHint: 'Pega el texto completo de tu CV',
      targetJob: 'Puesto Objetivo',
      targetJobPlaceholder: 'ej.: Gerente de Producto, Ingeniero Frontend',
      jobDesc: 'Descripción del Puesto (Opcional)',
      jobDescPlaceholder: 'Pega la descripción del puesto para una optimización dirigida...',
      optModules: 'Módulos de Optimización',
      optModulesHint: 'Elige qué optimizar',
      module_ats: 'Coincidencia de Palabras Clave ATS',
      module_ats_desc: 'Extraer y coincidir palabras clave ATS del puesto objetivo',
      module_star: 'Reescritura con Método STAR',
      module_star_desc: 'Reescribir experiencia laboral usando el método STAR',
      module_quant: 'Cuantificar Logros',
      module_quant_desc: 'Agregar datos y métricas para destacar logros',
      module_summary: 'Optimización de Resumen',
      module_summary_desc: 'Optimizar resumen profesional y competencias clave',
      module_format: 'Formato y Diseño',
      module_format_desc: 'Optimizar estructura y formato del CV',
      selectAll: 'Seleccionar Todo',
      deselectAll: 'Deseleccionar Todo',
      optimizing: 'IA Optimizando...',
      optimizeBtn: '✨ Optimizar Ahora',
      previewTab: 'CV Original',
      resultTab: 'Resultado Optimizado',
      defaultName: 'Tu Nombre',
      defaultSummary: 'Tu resumen aparecerá aquí...',
      atsScore: 'Puntuación ATS',
      keywords: 'Palabras Clave Recomendadas',
      suggestions: 'Sugerencias de Optimización',
      exportPdf: '📥 Exportar PDF',
      emptyResult: 'Pega tu CV y haz clic en Optimizar',
      pasteFirst: 'Primero pega el contenido de tu CV',
      optimized: '¡Optimización Completa!',
      optimizedTime: 'Tiempo',
      orUploadResume: 'O sube un archivo de CV',
      uploadFile: 'Subir Archivo',
      uploadHint: 'Soporta TXT, PDF, DOC, DOCX, máx 5MB',
      dragDrop: 'Arrastra y suelta o haz clic para subir',
      uploading: 'Subiendo...',
      uploadSuccess: 'Archivo subido correctamente',
      uploadError: 'Error al subir, intenta de nuevo',
      fetchJobUrl: 'Obtener desde URL',
      jobUrlPlaceholder: 'Pega la URL de la oferta, ej: https://...',
      fetching: 'Obteniendo...',
      fetchSuccess: 'Contenido de la oferta obtenido',
      fetchError: 'Error al obtener, verifica la URL',
      orPasteUrl: 'O pega el enlace de la oferta'
    },
    templates: {
      title: 'Plantillas de CV Profesionales',
      subtitle: 'Elige una plantilla que se adapte a tu sector',
      items: [
        { id: 'professional', name: 'Profesional', desc: 'Para puestos tradicionales y empresariales' },
        { id: 'modern', name: 'Moderno', desc: 'Para puestos tecnológicos' },
        { id: 'creative', name: 'Creativo', desc: 'Para puestos de diseño y creativos' },
        { id: 'academic', name: 'Académico', desc: 'Para puestos educativos e investigación' },
        { id: 'executive', name: 'Ejecutivo', desc: 'Para puestos de alta dirección' },
        { id: 'minimal', name: 'Minimalista', desc: 'Limpio y versátil' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: 'Crear CV',
      templates: 'Plantillas'
    },
    auth: {
      login: 'Iniciar sesión',
      register: 'Registrarse',
      email: 'Correo electrónico',
      password: 'Contraseña',
      name: 'Nombre',
      loginBtn: 'Iniciar sesión',
      registerBtn: 'Registrarse',
      noAccount: '¿No tienes cuenta?',
      hasAccount: '¿Ya tienes cuenta?'
    },
    jobs: {
      title: 'Ofertas de Empleo',
      subtitle: 'Descubre las últimas oportunidades laborales',
      search: 'Buscar empleos, empresas o ubicaciones...',
      apply: 'Aplicar',
      fullTime: 'Tiempo completo',
      intern: 'Pasantía'
    },
    generate: {
      title: 'Generador de CV desde Cero',
      subtitle: 'La IA te guía paso a paso para crear un CV profesional',
      placeholder: 'Escribe tu información...',
      send: 'Enviar'
    },
    perspective: {
      title: 'Análisis de 4 Perspectivas',
      original: 'Yo Original',
      optimized: 'Yo Optimizado',
      imagined: 'Yo Imaginado',
      desired: 'Lo que HR Quiere'
    },
    pricing: {
      title: 'Planes',
      subtitle: 'Elige el plan que mejor se adapte a ti',
      free: 'Gratis',
      pro: 'Pro',
      enterprise: 'Empresa',
      monthly: '/mes',
      current: 'Plan Actual',
      upgrade: 'Mejorar',
      popular: 'Más Popular'
    },
    faq: [
      { q: '¿Qué es ResumeTake?', a: 'ResumeTake es una herramienta de optimización de CV con IA que ayuda a generar CVs profesionales, coincidir automáticamente con palabras clave ATS y mejorar las tasas de éxito en la búsqueda de empleo.' },
      { q: '¿Es ResumeTake gratuito?', a: 'Sí, ResumeTake ofrece funciones básicas gratuitas de optimización de CV con IA.' },
      { q: '¿Cómo usar la optimización de CV con IA?', a: 'Sube tu CV o completa tu información, ingresa tu puesto objetivo y haz clic en el botón de optimizar IA para obtener sugerencias.' }
    ]
  },
  pt: {
    meta: {
      title: 'Construtor de Currículo com IA - ResumeTake | Otimização Inteligente',
      description: 'Otimizador de currículo com IA. Gere currículos profissionais, combine palavras-chave ATS e aumente seu sucesso na busca por emprego. Grátis com exportação PDF.',
      keywords: 'currículo IA,otimizador de currículo,currículo ATS,criador de currículo grátis,editor de currículo online,currículo profissional',
      editorTitle: 'Editor de Currículo - ResumeTake | Otimização IA',
      editorDesc: 'Otimize seu currículo com IA, combine palavras-chave ATS e aumente seu sucesso profissional.',
      editorKeywords: 'editor de currículo,otimização IA de currículo,editor online,otimização ATS',
      templatesTitle: 'Modelos de Currículo - ResumeTake | Modelos Profissionais Grátis',
      templatesDesc: 'Modelos de currículo profissionais para todos os setores. Grátis.',
      templatesKeywords: 'modelos de currículo,modelos grátis,currículo profissional,design de currículo'
    },
    nav: {
      home: 'Início',
      start: 'Começar',
      optimize: 'Otimizar IA',
      generate: 'Gerar',
      jobs: 'Vagas',
      pricing: 'Preços',
      createResume: 'Criar Currículo',
      templates: 'Modelos',
      language: 'Idioma',
      login: 'Entrar',
      logout: 'Sair'
    },
    hero: {
      badge: 'Otimização Inteligente de Currículo com IA',
      title1: 'Deixe a IA construir seu',
      title2: 'Currículo Perfeito',
      subtitle: 'Envie seu currículo e a IA otimiza automaticamente o conteúdo, combina palavras-chave ATS e melhora sua taxa de aprovação. Gere um currículo profissional com um clique.',
      cta: 'Começar →',
      learnMore: 'Saiba mais'
    },
    features: {
      title: 'Por que escolher o ResumeTake',
      subtitle: 'Powered by IA avançada para tornar a otimização de currículos simples e eficaz',
      items: [
        { icon: '⚡', title: 'Otimização IA Inteligente', desc: 'Otimiza automaticamente conteúdo, redação e formato do currículo para destacar competências-chave' },
        { icon: '🔍', title: 'Correspondência ATS', desc: 'Analisa inteligentemente vagas e combina automaticamente palavras-chave ATS' },
        { icon: '📄', title: 'Modelos Profissionais', desc: 'Múltiplos modelos para cada setor, troca de estilo com um clique, exportação PDF' }
      ]
    },
    cta: {
      title: 'Comece a otimizar seu currículo',
      subtitle: 'Use nosso otimizador de currículo com IA grátis para aumentar seu sucesso profissional',
      button: 'Começar Grátis →'
    },
    editor: {
      title: 'Otimização IA de Currículo',
      subtitle: 'Cole seu currículo, escolha módulos de otimização e obtenha resultados profissionais',
      pasteResume: 'Colar Conteúdo do Currículo',
      pasteResumePlaceholder: 'Cole o conteúdo do seu currículo aqui...\n\nSuporta texto simples: informações pessoais, experiência, educação, habilidades, etc.',
      pasteResumeHint: 'Cole o texto completo do seu currículo',
      targetJob: 'Cargo Alvo',
      targetJobPlaceholder: 'ex.: Gerente de Produto, Engenheiro Frontend',
      jobDesc: 'Descrição da Vaga (Opcional)',
      jobDescPlaceholder: 'Cole a descrição da vaga para uma otimização direcionada...',
      optModules: 'Módulos de Otimização',
      optModulesHint: 'Escolha o que otimizar',
      module_ats: 'Correspondência ATS',
      module_ats_desc: 'Extrair e combinar palavras-chave ATS da vaga alvo',
      module_star: 'Reescrita STAR',
      module_star_desc: 'Reescrever experiência usando o método STAR',
      module_quant: 'Quantificar Conquistas',
      module_quant_desc: 'Adicionar dados e métricas para destacar conquistas',
      module_summary: 'Otimização de Resumo',
      module_summary_desc: 'Otimizar resumo profissional e competências-chave',
      module_format: 'Formato e Layout',
      module_format_desc: 'Otimizar estrutura e formato do currículo',
      selectAll: 'Selecionar Tudo',
      deselectAll: 'Desmarcar Tudo',
      optimizing: 'IA Otimizando...',
      optimizeBtn: '✨ Otimizar Agora',
      previewTab: 'Currículo Original',
      resultTab: 'Resultado Otimizado',
      defaultName: 'Seu Nome',
      defaultSummary: 'Seu resumo aparecerá aqui...',
      atsScore: 'Pontuação ATS',
      keywords: 'Palavras-chave Recomendadas',
      suggestions: 'Sugestões de Otimização',
      exportPdf: '📥 Exportar PDF',
      emptyResult: 'Cole o currículo e clique em Otimizar',
      pasteFirst: 'Cole o conteúdo do currículo primeiro',
      optimized: 'Otimização Concluída!',
      optimizedTime: 'Tempo',
      orUploadResume: 'Ou faça upload de um currículo',
      uploadFile: 'Upload de Arquivo',
      uploadHint: 'Suporta TXT, PDF, DOC, DOCX, máx 5MB',
      dragDrop: 'Arraste e solte ou clique para enviar',
      uploading: 'Enviando...',
      uploadSuccess: 'Arquivo enviado com sucesso',
      uploadError: 'Falha no envio, tente novamente',
      fetchJobUrl: 'Obter a partir da URL',
      jobUrlPlaceholder: 'Cole a URL da vaga, ex: https://...',
      fetching: 'Obtendo...',
      fetchSuccess: 'Conteúdo da vaga obtido',
      fetchError: 'Falha ao obter, verifique a URL',
      orPasteUrl: 'Ou cole o link da vaga'
    },
    templates: {
      title: 'Modelos de Currículo Profissionais',
      subtitle: 'Escolha um modelo que se adapte ao seu setor',
      items: [
        { id: 'professional', name: 'Profissional', desc: 'Para cargos tradicionais e empresariais' },
        { id: 'modern', name: 'Moderno', desc: 'Para cargos de tecnologia' },
        { id: 'creative', name: 'Criativo', desc: 'Para cargos de design e criativos' },
        { id: 'academic', name: 'Acadêmico', desc: 'Para cargos educacionais e de pesquisa' },
        { id: 'executive', name: 'Executivo', desc: 'Para cargos de alta diretoria' },
        { id: 'minimal', name: 'Minimalista', desc: 'Limpo e versátil' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: 'Criar Currículo',
      templates: 'Modelos'
    },
    auth: {
      login: 'Entrar',
      register: 'Cadastrar',
      email: 'E-mail',
      password: 'Senha',
      name: 'Nome',
      loginBtn: 'Entrar',
      registerBtn: 'Cadastrar',
      noAccount: 'Não tem conta?',
      hasAccount: 'Já tem conta?'
    },
    jobs: {
      title: 'Vagas de Emprego',
      subtitle: 'Descubra as últimas oportunidades',
      search: 'Buscar vagas, empresas ou localizações...',
      apply: 'Candidatar',
      fullTime: 'Integral',
      intern: 'Estágio'
    },
    generate: {
      title: 'Gerador de Currículo do Zero',
      subtitle: 'A IA guia você passo a passo para criar um currículo profissional',
      placeholder: 'Digite suas informações...',
      send: 'Enviar'
    },
    perspective: {
      title: 'Análise de 4 Perspectivas',
      original: 'Eu Original',
      optimized: 'Eu Otimizado',
      imagined: 'Eu Imaginado',
      desired: 'O que a RH Quer'
    },
    pricing: {
      title: 'Planos',
      subtitle: 'Escolha o plano ideal para você',
      free: 'Grátis',
      pro: 'Pro',
      enterprise: 'Empresa',
      monthly: '/mês',
      current: 'Plano Atual',
      upgrade: 'Fazer Upgrade',
      popular: 'Mais Popular'
    },
    faq: [
      { q: 'O que é o ResumeTake?', a: 'ResumeTake é uma ferramenta de otimização de currículo com IA que ajuda a gerar currículos profissionais, combinar palavras-chave ATS automaticamente e melhorar as taxas de sucesso na busca por emprego.' },
      { q: 'O ResumeTake é gratuito?', a: 'Sim, o ResumeTake oferece funcionalidades básicas gratuitas de otimização de currículo com IA.' },
      { q: 'Como usar a otimização de currículo com IA?', a: 'Envie seu currículo ou preencha suas informações, insira seu cargo alvo e clique no botão de otimizar IA para obter sugestões.' }
    ]
  },
  fr: {
    meta: {
      title: 'Créateur de CV avec IA - ResumeTake | Optimisation Intelligente de CV',
      description: 'Optimiseur de CV propulsé par IA. Générez des CV professionnels, correspondez aux mots-clés ATS et augmentez votre succès dans la recherche d\'emploi. Gratuit avec export PDF.',
      keywords: 'CV IA,optimiseur CV,CV ATS,créateur CV gratuit,éditeur CV en ligne,CV professionnel,recherche emploi',
      editorTitle: 'Éditeur de CV - ResumeTake | Optimisation IA de CV',
      editorDesc: 'Optimisez votre CV avec l\'IA, correspondez aux mots-clés ATS et augmentez votre succès professionnel.',
      editorKeywords: 'éditeur CV,optimisation IA CV,éditeur en ligne,optimisation ATS',
      templatesTitle: 'Modèles de CV - ResumeTake | Modèles Professionnels Gratuits',
      templatesDesc: 'Modèles de CV professionnels pour chaque secteur. Gratuits.',
      templatesKeywords: 'modèles CV,modèles CV gratuits,CV professionnel,design CV'
    },
    nav: {
      home: 'Accueil',
      start: 'Commencer',
      optimize: 'Optimiser IA',
      generate: 'Générer',
      jobs: 'Emplois',
      pricing: 'Tarifs',
      createResume: 'Créer un CV',
      templates: 'Modèles',
      language: 'Langue',
      login: 'Connexion',
      logout: 'Déconnexion'
    },
    hero: {
      badge: 'Optimisation Intelligente de CV avec IA',
      title1: 'Laissez l\'IA construire votre',
      title2: 'CV Parfait',
      subtitle: 'Téléversez votre CV et l\'IA optimise automatiquement le contenu, correspond aux mots-clés ATS et améliore votre taux de réussite. Générez un CV professionnel en un clic.',
      cta: 'Commencer →',
      learnMore: 'En savoir plus'
    },
    features: {
      title: 'Pourquoi choisir ResumeTake',
      subtitle: 'Propulsé par une IA avancée pour rendre l\'optimisation de CV simple et efficace',
      items: [
        { icon: '⚡', title: 'Optimisation IA Intelligente', desc: 'Optimise automatiquement le contenu, la rédaction et le format du CV pour mettre en avant les compétences clés' },
        { icon: '🔍', title: 'Correspondance Mots-clés ATS', desc: 'Analyse intelligemment les offres et correspond automatiquement aux mots-clés ATS' },
        { icon: '📄', title: 'Modèles Professionnels', desc: 'Modèles multiples pour chaque secteur, changement de style en un clic, export PDF' }
      ]
    },
    cta: {
      title: 'Commencez à optimiser votre CV',
      subtitle: 'Utilisez notre optimiseur de CV IA gratuit pour augmenter votre succès professionnel',
      button: 'Commencer Gratuitement →'
    },
    editor: {
      title: 'Optimisation IA de CV',
      subtitle: 'Collez votre CV, choisissez les modules d\'optimisation, obtenez des résultats professionnels',
      pasteResume: 'Coller le Contenu du CV',
      pasteResumePlaceholder: 'Collez le contenu de votre CV ici...\n\nPrend en charge le texte brut : informations personnelles, expérience, formation, compétences, etc.',
      pasteResumeHint: 'Collez le texte complet de votre CV',
      targetJob: 'Poste Cible',
      targetJobPlaceholder: 'ex.: Chef de Produit, Ingénieur Frontend',
      jobDesc: 'Description du Poste (Optionnel)',
      jobDescPlaceholder: 'Collez la description du poste pour une optimisation ciblée...',
      optModules: 'Modules d\'Optimisation',
      optModulesHint: 'Choisissez ce qu\'optimiser',
      module_ats: 'Correspondance Mots-clés ATS',
      module_ats_desc: 'Extraire et correspondre les mots-clés ATS du poste cible',
      module_star: 'Réécriture Méthode STAR',
      module_star_desc: 'Réécrire l\'expérience en utilisant la méthode STAR',
      module_quant: 'Quantifier les Réalisations',
      module_quant_desc: 'Ajouter des données et métriques pour mettre en avant les réalisations',
      module_summary: 'Optimisation du Résumé',
      module_summary_desc: 'Optimiser le résumé professionnel et les compétences clés',
      module_format: 'Format et Mise en Page',
      module_format_desc: 'Optimiser la structure et le format du CV',
      selectAll: 'Tout Sélectionner',
      deselectAll: 'Tout Décocher',
      optimizing: 'IA en cours...',
      optimizeBtn: '✨ Optimiser Maintenant',
      previewTab: 'CV Original',
      resultTab: 'Résultat Optimisé',
      defaultName: 'Votre Nom',
      defaultSummary: 'Votre résumé apparaîtra ici...',
      atsScore: 'Score de Correspondance ATS',
      keywords: 'Mots-clés Recommandés',
      suggestions: 'Suggestions d\'Optimisation',
      exportPdf: '📥 Exporter PDF',
      emptyResult: 'Collez votre CV et cliquez sur Optimiser',
      pasteFirst: 'Collez d\'abord le contenu de votre CV',
      optimized: 'Optimisation Terminée!',
      optimizedTime: 'Durée',
      orUploadResume: 'Ou téléchargez un fichier CV',
      uploadFile: 'Télécharger un fichier',
      uploadHint: 'Supporte TXT, PDF, DOC, DOCX, max 5 Mo',
      dragDrop: 'Glissez-déposez ou cliquez pour télécharger',
      uploading: 'Téléchargement...',
      uploadSuccess: 'Fichier téléchargé avec succès',
      uploadError: 'Échec du téléchargement, réessayez',
      fetchJobUrl: 'Récupérer depuis URL',
      jobUrlPlaceholder: "Collez l'URL de l'offre, ex: https://...",
      fetching: 'Récupération...',
      fetchSuccess: "Contenu de l'offre récupéré",
      fetchError: "Échec de la récupération, vérifiez l'URL",
      orPasteUrl: "Ou collez le lien de l'offre"
    },
    templates: {
      title: 'Modèles de CV Professionnels',
      subtitle: 'Choisissez un modèle adapté à votre secteur',
      items: [
        { id: 'professional', name: 'Professionnel', desc: 'Pour les postes traditionnels et d\'entreprise' },
        { id: 'modern', name: 'Moderne', desc: 'Pour les postes technologiques' },
        { id: 'creative', name: 'Créatif', desc: 'Pour les postes de design et créatifs' },
        { id: 'academic', name: 'Académique', desc: 'Pour les postes éducatifs et de recherche' },
        { id: 'executive', name: 'Exécutif', desc: 'Pour les postes de haute direction' },
        { id: 'minimal', name: 'Minimaliste', desc: 'Épuré et polyvalent' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: 'Créer un CV',
      templates: 'Modèles'
    },
    auth: {
      login: 'Connexion',
      register: 'Inscription',
      email: 'E-mail',
      password: 'Mot de passe',
      name: 'Nom',
      loginBtn: 'Se connecter',
      registerBtn: "S'inscrire",
      noAccount: "Vous n'avez pas de compte ?",
      hasAccount: 'Vous avez déjà un compte ?'
    },
    jobs: {
      title: "Offres d'Emploi",
      subtitle: 'Découvrez les dernières opportunités',
      search: 'Rechercher emplois, entreprises ou lieux...',
      apply: 'Postuler',
      fullTime: 'Temps plein',
      intern: 'Stage'
    },
    generate: {
      title: 'Générateur de CV depuis Zéro',
      subtitle: "L'IA vous guide étape par étape pour créer un CV professionnel",
      placeholder: 'Entrez vos informations...',
      send: 'Envoyer'
    },
    perspective: {
      title: 'Analyse à 4 Perspectives',
      original: 'Moi Original',
      optimized: 'Moi Optimisé',
      imagined: 'Moi Imaginé',
      desired: 'Ce que HR Veut'
    },
    pricing: {
      title: 'Tarifs',
      subtitle: 'Choisissez le forfait qui vous convient le mieux',
      free: 'Gratuit',
      pro: 'Pro',
      enterprise: 'Entreprise',
      monthly: '/mois',
      current: 'Forfait Actuel',
      upgrade: 'Passer au Supérieur',
      popular: 'Le Plus Populaire'
    },
    faq: [
      { q: 'Qu\'est-ce que ResumeTake ?', a: 'ResumeTake est un outil d\'optimisation de CV propulsé par l\'IA qui aide à générer des CV professionnels, à correspondre automatiquement aux mots-clés ATS et à améliorer les taux de réussite dans la recherche d\'emploi.' },
      { q: 'ResumeTake est-il gratuit ?', a: 'Oui, ResumeTake propose des fonctionnalités gratuites de base d\'optimisation de CV avec l\'IA.' },
      { q: 'Comment utiliser l\'optimisation de CV avec l\'IA ?', a: 'Téléversez votre CV ou remplissez vos informations, entrez votre poste cible et cliquez sur le bouton d\'optimisation IA pour obtenir des suggestions.' }
    ]
  },
  de: {
    meta: {
      title: 'KI-Lebenslauf-Editor - ResumeTake | Intelligente Lebenslauf-Optimierung',
      description: 'KI-gestützter Lebenslauf-Optimierer. Erstellen Sie professionelle Lebensläufe, passen Sie ATS-Schlüsselwörter an und steigern Sie Ihren Erfolg bei der Jobsuche. Kostenlos mit PDF-Export.',
      keywords: 'KI Lebenslauf,Lebenslauf Optimierer,ATS Lebenslauf,kostenloser Lebenslauf Editor,online Lebenslauf Editor,professioneller Lebenslauf',
      editorTitle: 'Lebenslauf-Editor - ResumeTake | KI-Lebenslauf-Optimierung',
      editorDesc: 'Optimieren Sie Ihren Lebenslauf mit KI, passen Sie ATS-Schlüsselwörter an und steigern Sie Ihren beruflichen Erfolg.',
      editorKeywords: 'Lebenslauf Editor,KI Lebenslauf Optimierung,online Editor,ATS Optimierung',
      templatesTitle: 'Lebenslauf-Vorlagen - ResumeTake | Kostenlose Professionelle Vorlagen',
      templatesDesc: 'Professionelle Lebenslauf-Vorlagen für jede Branche. Kostenlos.',
      templatesKeywords: 'Lebenslauf Vorlagen,kostenlose Vorlagen,professioneller Lebenslauf,Lebenslauf Design'
    },
    nav: {
      home: 'Startseite',
      start: 'Jetzt starten',
      optimize: 'KI-Optimierung',
      generate: 'Generieren',
      jobs: 'Stellenangebote',
      pricing: 'Preise',
      createResume: 'Lebenslauf erstellen',
      templates: 'Vorlagen',
      language: 'Sprache',
      login: 'Anmelden',
      logout: 'Abmelden'
    },
    hero: {
      badge: 'KI-gestützte intelligente Lebenslauf-Optimierung',
      title1: 'Lassen Sie KI Ihren',
      title2: 'perfekten Lebenslauf erstellen',
      subtitle: 'Laden Sie Ihren Lebenslauf hoch und die KI optimiert automatisch den Inhalt, passt ATS-Schlüsselwörter an und verbessert Ihre Erfolgsquote. Erstellen Sie mit einem Klick einen professionellen Lebenslauf.',
      cta: 'Jetzt starten →',
      learnMore: 'Mehr erfahren'
    },
    features: {
      title: 'Warum ResumeTake wählen',
      subtitle: 'Unterstützt von fortschrittlicher KI-Technologie für einfache und effektive Lebenslauf-Optimierung',
      items: [
        { icon: '⚡', title: 'KI-Smart-Optimierung', desc: 'Optimiert automatisch Lebenslauf-Inhalt, Formulierung und Formatierung zur Hervorhebung der Kernkompetenzen' },
        { icon: '🔍', title: 'ATS-Schlüsselwort-Matching', desc: 'Analysiert intelligenterweise Stellenanzeigen und passt automatisch ATS-Schlüsselwörter an' },
        { icon: '📄', title: 'Professionelle Vorlagen', desc: 'Mehrere Vorlagen für jede Branche, One-Click-Stilwechsel, PDF-Export' }
      ]
    },
    cta: {
      title: 'Beginnen Sie mit der Optimierung Ihres Lebenslaufs',
      subtitle: 'Nutzen Sie unseren kostenlosen KI-Lebenslauf-Optimierer für mehr Erfolg bei der Jobsuche',
      button: 'Kostenlos starten →'
    },
    editor: {
      title: 'KI-Lebenslauf-Optimierung',
      subtitle: 'Lebenslauf einfügen, Optimierungsmodul wählen, professionelle Ergebnisse erhalten',
      pasteResume: 'Lebenslauf einfügen',
      pasteResumePlaceholder: 'Fügen Sie hier Ihren Lebenslauf ein...\n\nKlartext unterstützt: persönliche Daten, Berufserfahrung, Bildung, Fähigkeiten usw.',
      pasteResumeHint: 'Fügen Sie den gesamten Lebenslauftext ein',
      targetJob: 'Zielposition',
      targetJobPlaceholder: 'z.B.: Produktmanager, Frontend-Entwickler',
      jobDesc: 'Stellenbeschreibung (Optional)',
      jobDescPlaceholder: 'Stellenbeschreibung einfügen für gezielte Optimierung...',
      optModules: 'Optimierungsmodul',
      optModulesHint: 'Wählen Sie was optimiert werden soll',
      module_ats: 'ATS-Schlüsselwort-Matching',
      module_ats_desc: 'ATS-Schlüsselwörter der Zielposition extrahieren und abgleichen',
      module_star: 'STAR-Methode',
      module_star_desc: 'Berufserfahrung mit STAR-Methode umschreiben',
      module_quant: 'Erfolge quantifizieren',
      module_quant_desc: 'Daten und Kennzahlen hinzufügen',
      module_summary: 'Profil-Optimierung',
      module_summary_desc: 'Berufsprofil und Kernkompetenzen optimieren',
      module_format: 'Format & Layout',
      module_format_desc: 'Struktur und Format des Lebenslaufs optimieren',
      selectAll: 'Alle Auswählen',
      deselectAll: 'Alle Abwählen',
      optimizing: 'KI optimiert...',
      optimizeBtn: '✨ Jetzt Optimieren',
      previewTab: 'Original-Lebenslauf',
      resultTab: 'Optimiertes Ergebnis',
      defaultName: 'Ihr Name',
      defaultSummary: 'Ihr Profil wird hier angezeigt...',
      atsScore: 'ATS-Übereinstimmung',
      keywords: 'Empfohlene Schlüsselwörter',
      suggestions: 'Optimierungsvorschläge',
      exportPdf: '📥 PDF exportieren',
      emptyResult: 'Lebenslauf einfügen und auf Optimieren klicken',
      pasteFirst: 'Bitte zuerst Lebenslauf einfügen',
      optimized: 'Optimierung Abgeschlossen!',
      optimizedTime: 'Dauer',
      orUploadResume: 'Oder einen Lebenslauf hochladen',
      uploadFile: 'Datei hochladen',
      uploadHint: 'Unterstützt TXT, PDF, DOC, DOCX, max 5MB',
      dragDrop: 'Drag & Drop oder Klicken zum Hochladen',
      uploading: 'Hochladen...',
      uploadSuccess: 'Datei erfolgreich hochgeladen',
      uploadError: 'Upload fehlgeschlagen, bitte erneut versuchen',
      fetchJobUrl: 'Von URL abrufen',
      jobUrlPlaceholder: 'Stellenangebot-URL einfügen, z.B. https://...',
      fetching: 'Abrufen...',
      fetchSuccess: 'Stelleninhalt abgerufen',
      fetchError: 'Abrufen fehlgeschlagen, URL überprüfen',
      orPasteUrl: 'Oder Stellenangebot-Link einfügen'
    },
    templates: {
      title: 'Professionelle Lebenslauf-Vorlagen',
      subtitle: 'Wählen Sie eine Vorlage für Ihre Branche',
      items: [
        { id: 'professional', name: 'Professionell', desc: 'Für traditionelle und Geschäftspositionen' },
        { id: 'modern', name: 'Modern', desc: 'Für Tech- und Startup-Positionen' },
        { id: 'creative', name: 'Kreativ', desc: 'Für Design- und Kreativpositionen' },
        { id: 'academic', name: 'Akademisch', desc: 'Für Bildungs- und Forschungspositionen' },
        { id: 'executive', name: 'Führungskraft', desc: 'Für Senior-Management-Positionen' },
        { id: 'minimal', name: 'Minimalistisch', desc: 'Aufgerüstet und vielseitig' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: 'Lebenslauf erstellen',
      templates: 'Vorlagen'
    },
    auth: {
      login: 'Anmelden',
      register: 'Registrieren',
      email: 'E-Mail',
      password: 'Passwort',
      name: 'Name',
      loginBtn: 'Anmelden',
      registerBtn: 'Registrieren',
      noAccount: 'Noch kein Konto?',
      hasAccount: 'Bereits ein Konto?'
    },
    jobs: {
      title: 'Stellenangebote',
      subtitle: 'Entdecken Sie die neuesten Jobmöglichkeiten',
      search: 'Jobs, Unternehmen oder Standorte suchen...',
      apply: 'Bewerben',
      fullTime: 'Vollzeit',
      intern: 'Praktikum'
    },
    generate: {
      title: 'Lebenslauf-Generator von Null',
      subtitle: 'KI begleitet Sie Schritt für Schritt zum professionellen Lebenslauf',
      placeholder: 'Geben Sie Ihre Informationen ein...',
      send: 'Senden'
    },
    perspective: {
      title: '4-Perspektiven-Analyse',
      original: 'Ich (Original)',
      optimized: 'Ich (Optimiert)',
      imagined: 'Ich (Vorgestellt)',
      desired: 'Was HR Will'
    },
    pricing: {
      title: 'Preise',
      subtitle: 'Wählen Sie den passenden Plan',
      free: 'Kostenlos',
      pro: 'Pro',
      enterprise: 'Unternehmen',
      monthly: '/Monat',
      current: 'Aktueller Plan',
      upgrade: 'Upgrades',
      popular: 'Beliebteste'
    },
    faq: [
      { q: 'Was ist ResumeTake?', a: 'ResumeTake ist ein KI-gestütztes Lebenslauf-Optimierungstool, das bei der Erstellung professioneller Lebensläufe hilft, ATS-Schlüsselwörter automatisch abzugleichen und die Erfolgsquote bei der Jobsuche zu steigern.' },
      { q: 'Ist ResumeTake kostenlos?', a: 'Ja, ResumeTake bietet grundlegende kostenlose KI-Lebenslauf-Optimierungsfunktionen.' },
      { q: 'Wie nutzt man die KI-Lebenslauf-Optimierung?', a: 'Laden Sie Ihren Lebenslauf hoch oder füllen Sie Ihre Daten aus, geben Sie Ihre Zielposition ein und klicken Sie auf die KI-Optimierungsschaltfläche, um Vorschläge zu erhalten.' }
    ]
  },
  ar: {
    meta: {
      title: 'منشئ السيرة الذاتية بالذكاء الاصطناعي - ResumeTake | تحسين ذكي للسيرة الذاتية',
      description: 'محسّن السيرة الذاتية بالذكاء الاصطناعي. أنشئ سير ذاتية احترافية، وطابق كلمات ATS المفتاحية، وازد	success في البحث عن عمل. مجاني مع تصدير PDF.',
      keywords: 'سيرة ذاتية ذكاء اصطناعي,تحسين سيرة ذاتية,ATS سيرة ذاتية,منشئ سيرة الذاتية المجاني,محرر سيرة الذاتية عبر الإنترنت',
      editorTitle: 'محرر السيرة الذاتية - ResumeTake | تحسين بالذكاء الاصطناعي',
      editorDesc: 'حسّن سيرتك الذاتية بالذكاء الاصطناعي وطابق كلمات ATS المفتاحية لزيادة نجاحك.',
      editorKeywords: 'محرر السيرة الذاتية,تحسين بالذكاء الاصطناعي,محرر عبر الإنترنت,تحسين ATS',
      templatesTitle: 'قوالب السيرة الذاتية - ResumeTake | قوالب احترافية مجانية',
      templatesDesc: 'قوالب سيرة ذاتية احترافية لكل صناعة. مجانية.',
      templatesKeywords: 'قوالب السيرة الذاتية,قوالب مجانية,سيرة ذاتية احترافية,تصميم سيرة ذاتية'
    },
    nav: {
      home: 'الرئيسية',
      start: 'ابدأ مجاناً',
      optimize: 'تحسين بالذكاء',
      generate: 'توليد',
      jobs: 'الوظائف',
      pricing: 'الأسعار',
      createResume: 'إنشاء سيرة ذاتية',
      templates: 'القوالب',
      language: 'اللغة',
      login: 'تسجيل الدخول',
      logout: 'تسجيل الخروج'
    },
    hero: {
      badge: 'تحسين ذكي للسيرة الذاتية بالذكاء الاصطناعي',
      title1: 'دع الذكاء الاصطناعي يبني',
      title2: 'سيرتك الذاتية المثالية',
      subtitle: 'ارفع سيرتك الذاتية وسيقوم الذكاء الاصطناعي تلقائياً بتحسين المحتوى ومطابقة كلمات ATS المفتاحية وزيادة معدل النجاح. أنشئ سيرة ذاتية احترافية بنقرة واحدة.',
      cta: 'ابدأ مجاناً →',
      learnMore: 'اعرف المزيد'
    },
    features: {
      title: 'لماذا تختار ResumeTake',
      subtitle: 'مدعوم بتقنية ذكاء اصطناعي متقدمة لجعل تحسين السيرة الذاتية بسيطاً وفعالاً',
      items: [
        { icon: '⚡', title: 'تحسين ذكي بالذكاء الاصطناعي', desc: 'حسّن تلقائياً محتوى وصياغة وتنسيق السيرة الذاتية لإبراز الكفاءات الأساسية' },
        { icon: '🔍', title: 'مطابقة كلمات ATS المفتاحية', desc: 'حلل بذكاء الوظائف ومطابق تلقائياً كلمات ATS المفتاحية' },
        { icon: '📄', title: 'قوالب احترافية', desc: 'قوالب متعددة لكل صناعة، تبديل ستايل بنقرة واحدة، تصدير PDF' }
      ]
    },
    cta: {
      title: 'ابدأ بتحسين سيرتك الذاتية',
      subtitle: 'استخدم محسّن السيرة الذاتية المجاني بالذكاء الاصطناعي لزيادة نجاحك',
      button: 'ابدأ مجاناً →'
    },
    editor: {
      title: 'تحسين السيرة الذاتية بالذكاء الاصطناعي',
      subtitle: 'الصق سيرتك الذاتية، واختر وحدات التحسين، واحصل على نتائج احترافية',
      pasteResume: 'لصق محتوى السيرة الذاتية',
      pasteResumePlaceholder: 'الصق محتوى سيرتك الذاتية هنا...\n\nيدعم النص العادي: المعلومات الشخصية، الخبرات العملية، التعليم، المهارات، إلخ',
      pasteResumeHint: 'الصق نص سيرتك الذاتية الكامل',
      targetJob: 'المنصب المستهدف',
      targetJobPlaceholder: 'مثال: مدير المنتجات، مهندس الواجهة الأمامية',
      jobDesc: 'وصف الوظيفة (اختياري)',
      jobDescPlaceholder: 'الصق وصف الوظيفة لتحسين مستهدف...',
      optModules: 'وحدات التحسين',
      optModulesHint: 'اختر ما تريد تحسينه',
      module_ats: 'مطابقة كلمات ATS',
      module_ats_desc: 'استخراج ومطابقة كلمات ATS للمنصب المستهدف',
      module_star: 'إعادة الصياغة بطريقة STAR',
      module_star_desc: 'إعادة كتابة الخبرات بطريقة STAR',
      module_quant: 'تقدير الإنجازات',
      module_quant_desc: 'إضافة البيانات والمقاييس لإبراز الإنجازات',
      module_summary: 'تحسين الملخص المهني',
      module_summary_desc: 'تحسين الملخص المهني والكفاءات الأساسية',
      module_format: 'التنسيق والتخطيط',
      module_format_desc: 'تحسين هيكل وتنسيق السيرة الذاتية',
      selectAll: 'تحديد الكل',
      deselectAll: 'إلغاء التحديد',
      optimizing: 'الذكاء الاصطناعي يحسّن...',
      optimizeBtn: '✨ حسّن الآن',
      previewTab: 'السيرة الذاتية الأصلية',
      resultTab: 'النتيجة المحسّنة',
      defaultName: 'اسمك',
      defaultSummary: 'سيظهر ملخصك هنا...',
      atsScore: 'نتيجة مطابقة ATS',
      keywords: 'الكلمات المفتاحية المقترحة',
      suggestions: 'اقتراحات التحسين',
      exportPdf: '📥 تصدير PDF',
      emptyResult: 'الصق السيرة الذاتية واضغط تحسين',
      pasteFirst: 'الصق محتوى السيرة الذاتية أولاً',
      optimized: 'اكتمل التحسين!',
      optimizedTime: 'الوقت',
      orUploadResume: 'أو قم بتحميل ملف السيرة الذاتية',
      uploadFile: 'تحميل ملف',
      uploadHint: 'يدعم TXT, PDF, DOC, DOCX، حد أقصى 5 ميجا',
      dragDrop: 'اسحب وأفلت أو انقر للتحميل',
      uploading: 'جاري التحميل...',
      uploadSuccess: 'تم تحميل الملف بنجاح',
      uploadError: 'فشل التحميل، حاول مرة أخرى',
      fetchJobUrl: 'جلب من الرابط',
      jobUrlPlaceholder: 'الصق رابط صفحة الوظيفة، مثال: https://...',
      fetching: 'جاري الجلب...',
      fetchSuccess: 'تم جلب محتوى الوظيفة',
      fetchError: 'فشل الجلب، تحقق من الرابط',
      orPasteUrl: 'أو الصق رابط الوظيفة'
    },
    templates: {
      title: 'قوالب سيرة ذاتية احترافية',
      subtitle: 'اختر قالباً يناسب صناعتك',
      items: [
        { id: 'professional', name: 'احترافي', desc: 'للمناصب التقليدية والتجارية' },
        { id: 'modern', name: 'عصري', desc: 'للمناصب التقنية' },
        { id: 'creative', name: 'إبداعي', desc: 'لمناصب التصميم والإبداع' },
        { id: 'academic', name: 'أكاديمي', desc: 'لمناصب التعليم والبحث' },
        { id: 'executive', name: 'تنفيذي', desc: 'لمناصب الإدارة العليا' },
        { id: 'minimal', name: 'بسيط', desc: 'أنيق وعملي' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: 'إنشاء سيرة ذاتية',
      templates: 'القوالب'
    },
    auth: {
      login: 'تسجيل الدخول',
      register: 'إنشاء حساب',
      email: 'البريد الإلكتروني',
      password: 'كلمة المرور',
      name: 'الاسم',
      loginBtn: 'تسجيل الدخول',
      registerBtn: 'إنشاء حساب',
      noAccount: 'ليس لديك حساب؟',
      hasAccount: 'لديك حساب بالفعل؟'
    },
    jobs: {
      title: 'الوظائف',
      subtitle: 'اكتشف أحدث فرص العمل',
      search: 'ابحث عن وظائف أو شركات أو مواقع...',
      apply: 'تقديم',
      fullTime: 'دوام كامل',
      intern: 'تدريب'
    },
    generate: {
      title: 'إنشاء سيرة ذاتية من الصفر',
      subtitle: 'الذكاء الاصطناعي يرشدك خطوة بخطوة لبناء سيرة ذاتية احترافية',
      placeholder: 'أدخل معلوماتك...',
      send: 'إرسال'
    },
    perspective: {
      title: 'تحليل 4 مناظر',
      original: 'الأنا الأصلية',
      optimized: 'الأنا المحسّنة',
      imagined: 'الأنا المتخيّلة',
      desired: 'ما يريده الموارد البشرية'
    },
    pricing: {
      title: 'الأسعار',
      subtitle: 'اختر الخطة المناسبة لك',
      free: 'مجاني',
      pro: 'احترافي',
      enterprise: 'مؤسسات',
      monthly: '/شهر',
      current: 'الخطة الحالية',
      upgrade: 'ترقية',
      popular: 'الأكثر شعبية'
    },
    faq: [
      { q: 'ما هو ResumeTake؟', a: 'ResumeTake هو أداة تحسين السيرة الذاتية بالذكاء الاصطناعي التي تساعد في إنشاء سير ذاتية احترافية ومطابقة كلمات ATS المفتاحية تلقائياً لزيادة نجاح البحث عن عمل.' },
      { q: 'هل ResumeTake مجاني؟', a: 'نعم، يقدم ResumeTake ميزات مجانية أساسية لتحسين السيرة الذاتية بالذكاء الاصطناعي.' },
      { q: 'كيف تستخدم تحسين السيرة الذاتية بالذكاء الاصطناعي؟', a: 'ارفع سيرتك الذاتية أو أدخل معلوماتك، وأدخل المنصب المستهدف، ثم انقر على زر التحسين للحصول على اقتراحات.' }
    ]
  },
  hi: {
    meta: {
      title: 'AI रिज़्यूमे बिल्डर - ResumeTake | स्मार्ट रिज़्यूमे ऑप्टिमाइज़ेशन',
      description: 'AI-संचालित रिज़्यूमे ऑप्टिमाइज़र। पेशेवर रिज़्यूमे बनाएं, ATS कीवर्ड मैच करें, और नौकरी खोज में सफलता बढ़ाएं। मुफ्त में PDF एक्सपोर्ट।',
      keywords: 'AI रिज़्यूमे,रिज़्यूमे ऑप्टिमाइज़र,ATS रिज़्यूमे,मुफ्त रिज़्यूमे टूल,ऑनलाइन रिज़्यूमे एडिटर,पेशेवर रिज़्यूमे',
      editorTitle: 'रिज़्यूमे एडिटर - ResumeTake | AI रिज़्यूमे ऑप्टिमाइज़ेशन',
      editorDesc: 'AI के साथ अपना रिज़्यूमे ऑप्टिमाइज़ करें, ATS कीवर्ड मैच करें, और नौकरी सफलता बढ़ाएं।',
      editorKeywords: 'रिज़्यूमे एडिटर,AI ऑप्टिमाइज़ेशन,ऑनलाइन एडिटर,ATS ऑप्टिमाइज़ेशन',
      templatesTitle: 'रिज़्यूमे टेम्पलेट्स - ResumeTake | मुफ्त पेशेवर टेम्पलेट्स',
      templatesDesc: 'हर उद्योग के लिए पेशेवर रिज़्यूमे टेम्पलेट्स। मुफ्त।',
      templatesKeywords: 'रिज़्यूमे टेम्पलेट्स,मुफ्त टेम्पलेट्स,पेशेवर रिज़्यूमे,रिज़्यूमे डिज़ाइन'
    },
    nav: {
      home: 'होम',
      start: 'शुरू करें',
      optimize: 'AI ऑप्टिमाइज़',
      generate: 'जनरेट करें',
      jobs: 'नौकरियां',
      pricing: 'मूल्य निर्धारण',
      createResume: 'रिज़्यूमे बनाएं',
      templates: 'टेम्पलेट्स',
      language: 'भाषा',
      login: 'लॉगिन',
      logout: 'लॉगआउट'
    },
    hero: {
      badge: 'AI-संचालित स्मार्ट रिज़्यूमे ऑप्टिमाइज़ेशन',
      title1: 'AI को बनाने दें आपका',
      title2: 'परफेक्ट रिज़्यूमे',
      subtitle: 'अपना रिज़्यूमे अपलोड करें और AI स्वचालित रूप से सामग्री को ऑप्टिमाइज़ करता है, ATS कीवर्ड मैच करता है, और पास दर बढ़ाता है। एक क्लिक में पेशेवर रिज़्यूमे बनाएं।',
      cta: 'शुरू करें →',
      learnMore: 'और जानें'
    },
    features: {
      title: 'ResumeTake क्यों चुनें',
      subtitle: 'उन्नत AI तकनीक से रिज़्यूमे ऑप्टिमाइज़ेशन को सरल और प्रभावी बनाने के लिए',
      items: [
        { icon: '⚡', title: 'AI स्मार्ट ऑप्टिमाइज़ेशन', desc: 'स्वचालित रूप से रिज़्यूमे सामग्री, शब्दावली और प्रारूप को ऑप्टिमाइज़ करता है' },
        { icon: '🔍', title: 'ATS कीवर्ड मैचिंग', desc: 'नौकरी की पोस्टिंग का बुद्धिमानी से विश्लेषण करता है और ATS कीवर्ड स्वचालित मैच करता है' },
        { icon: '📄', title: 'पेशेवर टेम्पलेट्स', desc: 'हर उद्योग के लिए कई टेम्पलेट्स, एक क्लिक में स्टाइल बदलें, PDF एक्सपोर्ट' }
      ]
    },
    cta: {
      title: 'अपना रिज़्यूमे ऑप्टिमाइज़ करना शुरू करें',
      subtitle: 'नौकरी सफलता बढ़ाने के लिए हमारे मुफ्त AI रिज़्यूमे ऑप्टिमाइज़र का उपयोग करें',
      button: 'मुफ्त में शुरू करें →'
    },
    editor: {
      title: 'AI रिज़्यूमे ऑप्टिमाइज़ेशन',
      subtitle: 'रिज़्यूमे पेस्ट करें, ऑप्टिमाइज़ेशन मॉड्यूल चुनें, पेशेवर परिणाम प्राप्त करें',
      pasteResume: 'रिज़्यूमे सामग्री पेस्ट करें',
      pasteResumePlaceholder: 'यहां अपना रिज़्यूमे पेस्ट करें...\n\nप्लेन टेक्स्ट समर्थित: व्यक्तिगत जानकारी, कार्य अनुभव, शिक्षा, कौशल, आदि।',
      pasteResumeHint: 'अपना पूरा रिज़्यूमे टेक्स्ट पेस्ट करें',
      targetJob: 'लक्षित पद',
      targetJobPlaceholder: 'जैसे: प्रोडक्ट मैनेजर, फ्रंटएंड इंजीनियर',
      jobDesc: 'नौकरी विवरण (वैकल्पिक)',
      jobDescPlaceholder: 'लक्षित ऑप्टिमाइज़ेशन के लिए नौकरी विवरण पेस्ट करें...',
      optModules: 'ऑप्टिमाइज़ेशन मॉड्यूल',
      optModulesHint: 'क्या ऑप्टिमाइज़ करना है चुनें',
      module_ats: 'ATS कीवर्ड मैचिंग',
      module_ats_desc: 'लक्षित नौकरी के ATS कीवर्ड निकालें और मैच करें',
      module_star: 'STAR विधि रीराइट',
      module_star_desc: 'STAR विधि से कार्य अनुभव दोबारा लिखें',
      module_quant: 'उपलब्धियाँ मात्रात्मक बनाएं',
      module_quant_desc: 'डेटा और मेट्रिक्स जोड़कर उपलब्धियाँ उजागर करें',
      module_summary: 'सारांश ऑप्टिमाइज़ेशन',
      module_summary_desc: 'पेशेवर सारांश और मुख्य कौशल ऑप्टिमाइज़ करें',
      module_format: 'फ़ॉर्मेट और लेआउट',
      module_format_desc: 'रिज़्यूमे संरचना और फ़ॉर्मेट ऑप्टिमाइज़ करें',
      selectAll: 'सभी चुनें',
      deselectAll: 'सभी हटाएं',
      optimizing: 'AI ऑप्टिमाइज़ हो रहा है...',
      optimizeBtn: '✨ अभी ऑप्टिमाइज़ करें',
      previewTab: 'मूल रिज़्यूमे',
      resultTab: 'ऑप्टिमाइज़ परिणाम',
      defaultName: 'आपका नाम',
      defaultSummary: 'आपका सारांश यहां दिखाई देगा...',
      atsScore: 'ATS मैच स्कोर',
      keywords: 'अनुशंसित कीवर्ड',
      suggestions: 'ऑप्टिमाइज़ेशन सुझाव',
      exportPdf: '📥 PDF एक्सपोर्ट',
      emptyResult: 'रिज़्यूमे पेस्ट करें और ऑप्टिमाइज़ पर क्लिक करें',
      pasteFirst: 'पहले रिज़्यूमे सामग्री पेस्ट करें',
      optimized: 'ऑप्टिमाइज़ेशन पूर्ण!',
      optimizedTime: 'समय',
      orUploadResume: 'या रिज़्यूमे फ़ाइल अपलोड करें',
      uploadFile: 'फ़ाइल अपलोड करें',
      uploadHint: 'TXT, PDF, DOC, DOCX समर्थित, अधिकतम 5MB',
      dragDrop: 'फ़ाइल यहाँ खींचें या अपलोड के लिए क्लिक करें',
      uploading: 'अपलोड हो रहा है...',
      uploadSuccess: 'फ़ाइल सफलतापूर्वक अपलोड हुई',
      uploadError: 'अपलोड विफल, कृपया पुनः प्रयास करें',
      fetchJobUrl: 'URL से प्राप्त करें',
      jobUrlPlaceholder: 'नौकरी पोस्टिंग URL पेस्ट करें, जैसे https://...',
      fetching: 'प्राप्त हो रहा है...',
      fetchSuccess: 'नौकरी सामग्री प्राप्त हुई',
      fetchError: 'प्राप्त करने में विफल, URL जाँचें',
      orPasteUrl: 'या नौकरी लिंक पेस्ट करें'
    },
    templates: {
      title: 'पेशेवर रिज़्यूमे टेम्पलेट्स',
      subtitle: 'अपने उद्योग के अनुकूल टेम्पलेट चुनें',
      items: [
        { id: 'professional', name: 'पेशेवर', desc: 'पारंपरिक और व्यावसायिक पदों के लिए' },
        { id: 'modern', name: 'आधुनिक', desc: 'तकनीकी और स्टार्टअप पदों के लिए' },
        { id: 'creative', name: 'रचनात्मक', desc: 'डिज़ाइन और रचनात्मक पदों के लिए' },
        { id: 'academic', name: 'शैक्षणिक', desc: 'शिक्षा और अनुसंधान पदों के लिए' },
        { id: 'executive', name: 'कार्यकारी', desc: 'वरिष्ठ प्रबंधन पदों के लिए' },
        { id: 'minimal', name: 'न्यूनतम', desc: 'साफ और बहुमुखी' }
      ]
    },
    footer: {
      copyright: 'All rights reserved.',
      createResume: 'रिज़्यूमे बनाएं',
      templates: 'टेम्पलेट्स'
    },
    auth: {
      login: 'लॉग इन',
      register: 'रजिस्टर',
      email: 'ईमेल',
      password: 'पासवर्ड',
      name: 'नाम',
      loginBtn: 'लॉग इन',
      registerBtn: 'रजिस्टर',
      noAccount: 'खाता नहीं है?',
      hasAccount: 'पहले से खाता है?'
    },
    jobs: {
      title: 'नौकरी बोर्ड',
      subtitle: 'नवीनतम नौकरी के अवसर खोजें',
      search: 'नौकरियां, कंपनियां या स्थान खोजें...',
      apply: 'आवेदन',
      fullTime: 'पूर्णकालिक',
      intern: 'इंटर्नशिप'
    },
    generate: {
      title: 'शून्य से रिज़्यूमे जनरेटर',
      subtitle: 'AI आपको कदम-दर-कदम पेशेवर रिज़्यूमे बनाने में मार्गदर्शन करता है',
      placeholder: 'अपनी जानकारी टाइप करें...',
      send: 'भेजें'
    },
    perspective: {
      title: '4-दृष्टिकोण विश्लेषण',
      original: 'मूल मैं',
      optimized: 'अनुकूलित मैं',
      imagined: 'कल्पना मैं',
      desired: 'HR क्या चाहता है'
    },
    pricing: {
      title: 'मूल्य निर्धारण',
      subtitle: 'अपने लिए सही योजना चुनें',
      free: 'मुफ्त',
      pro: 'प्रो',
      enterprise: 'एंटरप्राइज़',
      monthly: '/माह',
      current: 'वर्तमान योजना',
      upgrade: 'अपग्रेड',
      popular: 'सबसे लोकप्रिय'
    },
    faq: [
      { q: 'ResumeTake क्या है?', a: 'ResumeTake एक AI-संचालित रिज़्यूमे ऑप्टिमाइज़ेशन टूल है जो पेशेवर रिज़्यूमे बनाने, ATS कीवर्ड स्वचालित मैच करने और नौकरी खोज सफलता दर बढ़ाने में मदद करता है।' },
      { q: 'क्या ResumeTake मुफ्त है?', a: 'हां, ResumeTake मुफ्त बुनियादी AI रिज़्यूमे ऑप्टिमाइज़ेशन सुविधाएं प्रदान करता है।' },
      { q: 'AI रिज़्यूमे ऑप्टिमाइज़ेशन कैसे उपयोग करें?', a: 'अपना रिज़्यूमे अपलोड करें या जानकारी भरें, अपना लक्षित पद दर्ज करें, और ऑप्टिमाइज़ बटन पर क्लिक करें।' }
    ]
  }
};

export function getTranslation(lang) {
  return translations[lang] || translations.en;
}
