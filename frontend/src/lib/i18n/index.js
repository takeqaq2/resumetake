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
      createResume: '创建简历',
      templates: '模板',
      language: '语言'
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
      title: '简历编辑器',
      subtitle: '填写信息，AI将自动优化内容',
      basicInfo: '基本信息',
      name: '姓名',
      namePlaceholder: '请输入姓名',
      email: '邮箱',
      emailPlaceholder: 'email@example.com',
      phone: '电话',
      phonePlaceholder: '13800138000',
      summary: '个人简介',
      summaryPlaceholder: '简要介绍你的专业背景...',
      targetJob: '目标职位',
      targetJobPlaceholder: '例如：产品经理、前端工程师',
      jobDesc: '职位描述（可选）',
      jobDescPlaceholder: '粘贴职位描述...',
      optimizing: 'AI优化中...',
      optimizeBtn: '✨ AI智能优化',
      previewTab: '预览简历',
      resultTab: '优化结果',
      defaultName: '你的姓名',
      defaultSummary: '个人简介将显示在这里...',
      atsScore: 'ATS匹配度',
      keywords: '推荐关键词',
      suggestions: '优化建议',
      exportPdf: '📥 导出PDF',
      emptyResult: '点击"AI智能优化"查看结果'
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
      createResume: 'Create Resume',
      templates: 'Templates',
      language: 'Language'
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
      title: 'Resume Editor',
      subtitle: 'Fill in your details, and AI will automatically optimize the content',
      basicInfo: 'Basic Information',
      name: 'Full Name',
      namePlaceholder: 'Enter your name',
      email: 'Email',
      emailPlaceholder: 'email@example.com',
      phone: 'Phone',
      phonePlaceholder: '+1 (555) 000-0000',
      summary: 'Professional Summary',
      summaryPlaceholder: 'Briefly describe your professional background...',
      targetJob: 'Target Position',
      targetJobPlaceholder: 'e.g., Product Manager, Frontend Engineer',
      jobDesc: 'Job Description (Optional)',
      jobDescPlaceholder: 'Paste the job description here...',
      optimizing: 'AI Optimizing...',
      optimizeBtn: '✨ AI Smart Optimize',
      previewTab: 'Preview Resume',
      resultTab: 'Optimization Result',
      defaultName: 'Your Name',
      defaultSummary: 'Your summary will appear here...',
      atsScore: 'ATS Match Score',
      keywords: 'Recommended Keywords',
      suggestions: 'Optimization Suggestions',
      exportPdf: '📥 Export PDF',
      emptyResult: 'Click "AI Smart Optimize" to see results'
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
      createResume: '履歴書作成',
      templates: 'テンプレート',
      language: '言語'
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
      title: '履歴書エディタ',
      subtitle: '情報を入力すると、AIが自動でコンテンツを最適化します',
      basicInfo: '基本情報',
      name: '氏名',
      namePlaceholder: 'お名前を入力',
      email: 'メール',
      emailPlaceholder: 'email@example.com',
      phone: '電話番号',
      phonePlaceholder: '090-0000-0000',
      summary: '自己PR',
      summaryPlaceholder: 'あなたの専門的な背景を簡潔に紹介...',
      targetJob: '希望職種',
      targetJobPlaceholder: '例：プロダクトマネージャー、フロントエンドエンジニア',
      jobDesc: '職務記述書（オプション）',
      jobDescPlaceholder: '職務記述書をここに貼り付け...',
      optimizing: 'AI最適化中...',
      optimizeBtn: '✨ AIスマート最適化',
      previewTab: '履歴書プレビュー',
      resultTab: '最適化結果',
      defaultName: 'お名前',
      defaultSummary: 'ここに自己PRが表示されます...',
      atsScore: 'ATSマッチングスコア',
      keywords: '推奨キーワード',
      suggestions: '最適化提案',
      exportPdf: '📥 PDFエクスポート',
      emptyResult: '「AIスマート最適化」をクリックして結果を表示'
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
      createResume: '이력서 만들기',
      templates: '템플릿',
      language: '언어'
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
      title: '이력서 편집기',
      subtitle: '정보를 입력하면 AI가 자동으로 내용을 최적화합니다',
      basicInfo: '기본 정보',
      name: '이름',
      namePlaceholder: '이름을 입력하세요',
      email: '이메일',
      emailPlaceholder: 'email@example.com',
      phone: '전화번호',
      phonePlaceholder: '010-0000-0000',
      summary: '자기소개',
      summaryPlaceholder: '전문적인 배경을 간략히 소개하세요...',
      targetJob: '희망 직종',
      targetJobPlaceholder: '예: 프로덕트 매니저, 프론트엔드 엔지니어',
      jobDesc: '직무 설명 (선택사항)',
      jobDescPlaceholder: '직무 설명을 여기에 붙여넣으세요...',
      optimizing: 'AI 최적화 중...',
      optimizeBtn: '✨ AI 스마트 최적화',
      previewTab: '이력서 미리보기',
      resultTab: '최적화 결과',
      defaultName: '이름',
      defaultSummary: '여기에 자기소개가 표시됩니다...',
      atsScore: 'ATS 매칭 점수',
      keywords: '추천 키워드',
      suggestions: '최적화 제안',
      exportPdf: '📥 PDF 내보내기',
      emptyResult: '"AI 스마트 최적화"를 클릭하여 결과 보기'
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
      createResume: 'Crear CV',
      templates: 'Plantillas',
      language: 'Idioma'
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
      title: 'Editor de CV',
      subtitle: 'Completa tu información y la IA optimizará automáticamente el contenido',
      basicInfo: 'Información Básica',
      name: 'Nombre Completo',
      namePlaceholder: 'Ingresa tu nombre',
      email: 'Correo Electrónico',
      emailPlaceholder: 'correo@ejemplo.com',
      phone: 'Teléfono',
      phonePlaceholder: '+34 600 000 000',
      summary: 'Resumen Profesional',
      summaryPlaceholder: 'Describe brevemente tu formación profesional...',
      targetJob: 'Puesto Objetivo',
      targetJobPlaceholder: 'ej.: Gerente de Producto, Ingeniero Frontend',
      jobDesc: 'Descripción del Puesto (Opcional)',
      jobDescPlaceholder: 'Pega la descripción del puesto aquí...',
      optimizing: 'IA Optimizando...',
      optimizeBtn: '✨ Optimización IA Inteligente',
      previewTab: 'Vista Previa',
      resultTab: 'Resultado de Optimización',
      defaultName: 'Tu Nombre',
      defaultSummary: 'Tu resumen aparecerá aquí...',
      atsScore: 'Puntuación ATS',
      keywords: 'Palabras Clave Recomendadas',
      suggestions: 'Sugerencias de Optimización',
      exportPdf: '📥 Exportar PDF',
      emptyResult: 'Haz clic en "Optimización IA Inteligente" para ver resultados'
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
      createResume: 'Criar Currículo',
      templates: 'Modelos',
      language: 'Idioma'
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
      title: 'Editor de Currículo',
      subtitle: 'Preencha suas informações e a IA otimizará automaticamente o conteúdo',
      basicInfo: 'Informações Básicas',
      name: 'Nome Completo',
      namePlaceholder: 'Digite seu nome',
      email: 'E-mail',
      emailPlaceholder: 'email@exemplo.com',
      phone: 'Telefone',
      phonePlaceholder: '+55 11 0000-0000',
      summary: 'Resumo Profissional',
      summaryPlaceholder: 'Descreva brevemente seu histórico profissional...',
      targetJob: 'Cargo Alvo',
      targetJobPlaceholder: 'ex.: Gerente de Produto, Engenheiro Frontend',
      jobDesc: 'Descrição da Vaga (Opcional)',
      jobDescPlaceholder: 'Cole a descrição da vaga aqui...',
      optimizing: 'IA Otimizando...',
      optimizeBtn: '✨ Otimização IA Inteligente',
      previewTab: 'Pré-visualizar',
      resultTab: 'Resultado da Otimização',
      defaultName: 'Seu Nome',
      defaultSummary: 'Seu resumo aparecerá aqui...',
      atsScore: 'Pontuação ATS',
      keywords: 'Palavras-chave Recomendadas',
      suggestions: 'Sugestões de Otimização',
      exportPdf: '📥 Exportar PDF',
      emptyResult: 'Clique em "Otimização IA Inteligente" para ver resultados'
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
      createResume: 'Créer un CV',
      templates: 'Modèles',
      language: 'Langue'
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
      title: 'Éditeur de CV',
      subtitle: 'Remplissez vos informations et l\'IA optimisera automatiquement le contenu',
      basicInfo: 'Informations de Base',
      name: 'Nom Complet',
      namePlaceholder: 'Entrez votre nom',
      email: 'E-mail',
      emailPlaceholder: 'email@exemple.com',
      phone: 'Téléphone',
      phonePlaceholder: '+33 6 00 00 00 00',
      summary: 'Résumé Professionnel',
      summaryPlaceholder: 'Décrivez brièvement votre parcours professionnel...',
      targetJob: 'Poste Cible',
      targetJobPlaceholder: 'ex.: Chef de Produit, Ingénieur Frontend',
      jobDesc: 'Description du Poste (Optionnel)',
      jobDescPlaceholder: 'Collez la description du poste ici...',
      optimizing: 'IA en cours d\'optimisation...',
      optimizeBtn: '✨ Optimisation IA Intelligente',
      previewTab: 'Aperçu du CV',
      resultTab: 'Résultat d\'Optimisation',
      defaultName: 'Votre Nom',
      defaultSummary: 'Votre résumé apparaîtra ici...',
      atsScore: 'Score de Correspondance ATS',
      keywords: 'Mots-clés Recommandés',
      suggestions: 'Suggestions d\'Optimisation',
      exportPdf: '📥 Exporter PDF',
      emptyResult: 'Cliquez sur "Optimisation IA Intelligente" pour voir les résultats'
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
      createResume: 'Lebenslauf erstellen',
      templates: 'Vorlagen',
      language: 'Sprache'
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
      title: 'Lebenslauf-Editor',
      subtitle: 'Füllen Sie Ihre Daten aus und die KI optimiert automatisch den Inhalt',
      basicInfo: 'Grundlegende Informationen',
      name: 'Vollständiger Name',
      namePlaceholder: 'Geben Sie Ihren Namen ein',
      email: 'E-Mail',
      emailPlaceholder: 'email@beispiel.com',
      phone: 'Telefon',
      phonePlaceholder: '+49 170 000 0000',
      summary: 'Berufsprofil',
      summaryPlaceholder: 'Beschreiben Sie kurz Ihren beruflichen Hintergrund...',
      targetJob: 'Zielposition',
      targetJobPlaceholder: 'z.B.: Produktmanager, Frontend-Entwickler',
      jobDesc: 'Stellenbeschreibung (Optional)',
      jobDescPlaceholder: 'Fügen Sie die Stellenbeschreibung hier ein...',
      optimizing: 'KI optimiert...',
      optimizeBtn: '✨ KI-Smart-Optimierung',
      previewTab: 'Vorschau',
      resultTab: 'Optimierungsergebnis',
      defaultName: 'Ihr Name',
      defaultSummary: 'Ihr Profil wird hier angezeigt...',
      atsScore: 'ATS-Übereinstimmung',
      keywords: 'Empfohlene Schlüsselwörter',
      suggestions: 'Optimierungsvorschläge',
      exportPdf: '📥 PDF exportieren',
      emptyResult: 'Klicken Sie auf "KI-Smart-Optimierung" für Ergebnisse'
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
      createResume: 'إنشاء سيرة ذاتية',
      templates: 'القوالب',
      language: 'اللغة'
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
      title: 'محرر السيرة الذاتية',
      subtitle: 'أدخل معلوماتك وسيقوم الذكاء الاصطناعي تلقائياً بتحسين المحتوى',
      basicInfo: 'المعلومات الأساسية',
      name: 'الاسم الكامل',
      namePlaceholder: 'أدخل اسمك',
      email: 'البريد الإلكتروني',
      emailPlaceholder: 'email@example.com',
      phone: 'الهاتف',
      phonePlaceholder: '+966 5X XXX XXXX',
      summary: 'الملخص المهني',
      summaryPlaceholder: 'وصف موجز لخلفيتك المهنية...',
      targetJob: 'المنصب المستهدف',
      targetJobPlaceholder: 'مثال: مدير المنتجات، مهندس الواجهة الأمامية',
      jobDesc: 'وصف الوظيفة (اختياري)',
      jobDescPlaceholder: 'الصق وصف الوظيفة هنا...',
      optimizing: 'الذكاء الاصطناعي يحسّن...',
      optimizeBtn: '✨ تحسين ذكي بالذكاء الاصطناعي',
      previewTab: 'معاينة السيرة الذاتية',
      resultTab: 'نتيجة التحسين',
      defaultName: 'اسمك',
      defaultSummary: 'سيظهر ملخصك هنا...',
      atsScore: 'نتيجة مطابقة ATS',
      keywords: 'الكلمات المفتاحية المقترحة',
      suggestions: 'اقتراحات التحسين',
      exportPdf: '📥 تصدير PDF',
      emptyResult: 'انقر على "تحسين ذكي بالذكاء الاصطناعي" لعرض النتائج'
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
      createResume: 'रिज़्यूमे बनाएं',
      templates: 'टेम्पलेट्स',
      language: 'भाषा'
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
      title: 'रिज़्यूमे एडिटर',
      subtitle: 'अपनी जानकारी भरें और AI स्वचालित रूप से सामग्री ऑप्टिमाइज़ करेगा',
      basicInfo: 'बुनियादी जानकारी',
      name: 'पूरा नाम',
      namePlaceholder: 'अपना नाम दर्ज करें',
      email: 'ईमेल',
      emailPlaceholder: 'email@example.com',
      phone: 'फ़ोन',
      phonePlaceholder: '+91 XXXXX XXXXX',
      summary: 'पेशेवर सारांश',
      summaryPlaceholder: 'अपनी पेशेवर पृष्ठभूमि का संक्षिप्त विवरण...',
      targetJob: 'लक्षित पद',
      targetJobPlaceholder: 'जैसे: प्रोडक्ट मैनेजर, फ्रंटएंड इंजीनियर',
      jobDesc: 'नौकरी विवरण (वैकल्पिक)',
      jobDescPlaceholder: 'नौकरी का विवरण यहां पेस्ट करें...',
      optimizing: 'AI ऑप्टिमाइज़ हो रहा है...',
      optimizeBtn: '✨ AI स्मार्ट ऑप्टिमाइज़ेशन',
      previewTab: 'रिज़्यूमे पूर्वावलोकन',
      resultTab: 'ऑप्टिमाइज़ेशन परिणाम',
      defaultName: 'आपका नाम',
      defaultSummary: 'आपका सारांश यहां दिखाई देगा...',
      atsScore: 'ATS मैच स्कोर',
      keywords: 'अनुशंसित कीवर्ड',
      suggestions: 'ऑप्टिमाइज़ेशन सुझाव',
      exportPdf: '📥 PDF एक्सपोर्ट',
      emptyResult: 'परिणाम देखने के लिए "AI स्मार्ट ऑप्टिमाइज़ेशन" पर क्लिक करें'
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
