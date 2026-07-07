<script>
  import { page } from '$app/stores';
  import { getTranslation } from '$lib/i18n/index.js';

  let lang = $derived($page.params.lang);
  let t = $derived(getTranslation(lang));

  let loading = $state('');
  let error = $state('');

  const pricingText = {
    zh: {
      choosePlan: '选择适合你的方案', checkoutError: '创建支付会话失败', networkError: '网络错误',
      freeFeatures: ['5次AI优化/月', '基础模板', '校招信息'],
      proFeatures: ['无限AI优化', '高级模板', '4视角分析', '0基础生成', '简历导出'],
      enterpriseFeatures: ['API接入', '团队管理', '自定义模板']
    },
    en: {
      choosePlan: 'Choose Your Plan', checkoutError: 'Failed to create checkout session', networkError: 'Network error',
      freeFeatures: ['5 AI optimizations/month', 'Basic templates', 'Campus recruitment info'],
      proFeatures: ['Unlimited AI optimizations', 'Premium templates', '4-perspective analysis', '0-basis resume generation', 'Resume export'],
      enterpriseFeatures: ['API access', 'Team management', 'Custom templates']
    },
    ja: {
      choosePlan: 'あなたに合うプランを選択', checkoutError: '決済セッションの作成に失敗しました', networkError: 'ネットワークエラー',
      freeFeatures: ['月5回のAI最適化', '基本テンプレート', '求人情報'],
      proFeatures: ['無制限AI最適化', 'プレミアムテンプレート', '4視点分析', 'ゼロから履歴書生成', '履歴書エクスポート'],
      enterpriseFeatures: ['API連携', 'チーム管理', 'カスタムテンプレート']
    },
    ko: {
      choosePlan: '나에게 맞는 플랜 선택', checkoutError: '결제 세션 생성에 실패했습니다', networkError: '네트워크 오류',
      freeFeatures: ['월 5회 AI 최적화', '기본 템플릿', '채용 정보'],
      proFeatures: ['무제한 AI 최적화', '프리미엄 템플릿', '4가지 관점 분석', '무경력 이력서 생성', '이력서 내보내기'],
      enterpriseFeatures: ['API 연동', '팀 관리', '맞춤 템플릿']
    },
    es: {
      choosePlan: 'Elige Tu Plan', checkoutError: 'No se pudo crear la sesión de pago', networkError: 'Error de red',
      freeFeatures: ['5 optimizaciones IA/mes', 'Plantillas básicas', 'Información de empleo'],
      proFeatures: ['Optimizaciones IA ilimitadas', 'Plantillas premium', 'Análisis de 4 perspectivas', 'Generación desde cero', 'Exportación de currículum'],
      enterpriseFeatures: ['Acceso API', 'Gestión de equipo', 'Plantillas personalizadas']
    },
    pt: {
      choosePlan: 'Escolha Seu Plano', checkoutError: 'Falha ao criar sessão de pagamento', networkError: 'Erro de rede',
      freeFeatures: ['5 otimizações IA/mês', 'Modelos básicos', 'Informações de vagas'],
      proFeatures: ['Otimizações IA ilimitadas', 'Modelos premium', 'Análise de 4 perspectivas', 'Geração do zero', 'Exportação de currículo'],
      enterpriseFeatures: ['Acesso API', 'Gestão de equipe', 'Modelos personalizados']
    },
    fr: {
      choosePlan: 'Choisissez Votre Offre', checkoutError: 'Échec de création de la session de paiement', networkError: 'Erreur réseau',
      freeFeatures: ['5 optimisations IA/mois', 'Modèles basiques', 'Informations emploi'],
      proFeatures: ['Optimisations IA illimitées', 'Modèles premium', 'Analyse en 4 perspectives', 'Génération depuis zéro', 'Export du CV'],
      enterpriseFeatures: ['Accès API', 'Gestion d’équipe', 'Modèles personnalisés']
    },
    de: {
      choosePlan: 'Wählen Sie Ihren Plan', checkoutError: 'Checkout-Sitzung konnte nicht erstellt werden', networkError: 'Netzwerkfehler',
      freeFeatures: ['5 KI-Optimierungen/Monat', 'Basisvorlagen', 'Jobinformationen'],
      proFeatures: ['Unbegrenzte KI-Optimierung', 'Premium-Vorlagen', '4-Perspektiven-Analyse', 'Lebenslauf von Grund auf', 'Lebenslauf exportieren'],
      enterpriseFeatures: ['API-Zugriff', 'Teamverwaltung', 'Eigene Vorlagen']
    },
    ar: {
      choosePlan: 'اختر الخطة المناسبة', checkoutError: 'فشل إنشاء جلسة الدفع', networkError: 'خطأ في الشبكة',
      freeFeatures: ['5 تحسينات بالذكاء الاصطناعي/شهر', 'قوالب أساسية', 'معلومات وظائف'],
      proFeatures: ['تحسينات غير محدودة', 'قوالب مميزة', 'تحليل من 4 زوايا', 'إنشاء سيرة من الصفر', 'تصدير السيرة الذاتية'],
      enterpriseFeatures: ['وصول API', 'إدارة الفريق', 'قوالب مخصصة']
    },
    hi: {
      choosePlan: 'अपना प्लान चुनें', checkoutError: 'चेकआउट सत्र बनाने में विफल', networkError: 'नेटवर्क त्रुटि',
      freeFeatures: ['5 AI अनुकूलन/माह', 'बेसिक टेम्पलेट', 'नौकरी जानकारी'],
      proFeatures: ['असीमित AI अनुकूलन', 'प्रीमियम टेम्पलेट', '4-दृष्टिकोण विश्लेषण', 'शून्य से रिज़्यूमे निर्माण', 'रिज़्यूमे निर्यात'],
      enterpriseFeatures: ['API एक्सेस', 'टीम प्रबंधन', 'कस्टम टेम्पलेट']
    }
  };

  let pt = $derived(pricingText[lang] || pricingText.en);

  let plans = $derived([
    {
      id: 'free',
      name: t.pricing.free,
      price: '0',
      period: '',
      features: pt.freeFeatures,
      cta: t.pricing.current,
      highlighted: false,
      disabled: true
    },
    {
      id: 'pro',
      name: t.pricing.pro,
      price: '¥29',
      period: t.pricing.monthly,
      features: pt.proFeatures,
      cta: t.pricing.upgrade,
      highlighted: true,
      disabled: false
    },
    {
      id: 'enterprise',
      name: t.pricing.enterprise,
      price: '¥99',
      period: t.pricing.monthly,
      features: pt.enterpriseFeatures,
      cta: t.pricing.upgrade,
      highlighted: false,
      disabled: false
    }
  ]);

  async function handleUpgrade(planId) {
    const token = localStorage.getItem('token');
    if (!token) {
      window.location.href = `/${lang}/auth`;
      return;
    }
    loading = planId;
    error = '';
    try {
      const res = await fetch('/api/v1/create-checkout-session', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': 'Bearer ' + token },
        body: JSON.stringify({ plan: planId })
      });
      const data = await res.json();
      if (data.url) {
        window.location.href = data.url;
      } else {
        error = data.message || pt.checkoutError;
      }
    } catch {
      error = pt.networkError;
    } finally {
      loading = '';
    }
  }
</script>

<svelte:head>
  <title>{t.pricing.title} - ResumeTake</title>
</svelte:head>

<div class="pricing-page">
  <div class="pricing-header">
    <div class="orb orb-blue animate-float" style="width:250px;height:250px;top:-15%;left:5%"></div>
    <div class="orb orb-purple animate-float" style="width:200px;height:200px;bottom:-10%;right:10%;animation-delay:2s"></div>
    <div class="container" style="position:relative;text-align:center">
      <span class="section-badge">💎 {t.pricing.title}</span>
      <h1 style="font-size:clamp(1.75rem,4vw,2.5rem);font-weight:800;margin:1rem 0 0.75rem">
        {pt.choosePlan}
      </h1>
      <p style="color:var(--text-secondary);font-size:1rem;max-width:32rem;margin:0 auto">{t.pricing.subtitle}</p>
    </div>
  </div>

  <div class="container" style="padding:3rem 1.5rem;margin-top:-2rem">
    {#if error}
      <div style="max-width:24rem;margin:0 auto 1.5rem;padding:0.75rem 1rem;border-radius:var(--radius);background:rgba(239,68,68,0.08);border:1px solid rgba(239,68,68,0.2);color:#ef4444;font-size:0.875rem;text-align:center">{error}</div>
    {/if}
    <div class="pricing-grid">
      {#each plans as plan}
        <div class="pricing-card {plan.highlighted ? 'highlighted' : ''}">
          {#if plan.highlighted}
            <div class="popular-badge">{t.pricing.popular}</div>
          {/if}
          <div class="plan-header">
            <h3 class="plan-name">{plan.name}</h3>
            <div class="plan-price">
              <span class="price-value">{plan.price}</span>
              {#if plan.period}
                <span class="price-period">/{plan.period}</span>
              {/if}
            </div>
          </div>
          <ul class="plan-features">
            {#each plan.features as feature}
              <li>
                <svg width="16" height="16" viewBox="0 0 16 16" fill="none"><path d="M3 8l3.5 3.5L13 5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                {feature}
              </li>
            {/each}
          </ul>
          <button class="btn {plan.highlighted ? 'btn-primary' : 'btn-secondary'} plan-cta" disabled={plan.disabled || loading === plan.id} onclick={() => handleUpgrade(plan.id)}>
            {#if loading === plan.id}
              <span class="auth-spinner"></span>
            {:else}
              {plan.cta}
            {/if}
          </button>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .pricing-page { overflow: hidden; }
  .pricing-header {
    position: relative; padding: 5rem 0 6rem;
    background: var(--gradient-hero); background-size: 200% 200%;
    animation: gradientShift 10s ease-in-out infinite;
  }
  .section-badge {
    display: inline-flex; align-items: center; gap: 0.375rem;
    padding: 0.375rem 0.875rem; border-radius: 9999px;
    background: var(--bg-glass); border: 1px solid var(--border);
    font-size: 0.8125rem; font-weight: 500; color: var(--primary);
    backdrop-filter: blur(8px);
  }
  .pricing-grid {
    display: grid; grid-template-columns: repeat(3, 1fr);
    gap: 1.5rem; align-items: stretch;
  }
  .pricing-card {
    background: var(--bg-glass); border: 1px solid var(--border);
    border-radius: var(--radius-lg); padding: 2rem;
    backdrop-filter: blur(16px); position: relative;
    display: flex; flex-direction: column;
    transition: all 0.35s cubic-bezier(0.4,0,0.2,1);
  }
  .pricing-card:hover {
    transform: translateY(-4px); box-shadow: var(--shadow-lg);
  }
  .pricing-card.highlighted {
    border-color: var(--primary);
    box-shadow: 0 8px 40px var(--primary-glow);
    transform: scale(1.02);
  }
  .pricing-card.highlighted:hover {
    transform: scale(1.02) translateY(-4px);
  }
  .popular-badge {
    position: absolute; top: -0.75rem; left: 50%; transform: translateX(-50%);
    padding: 0.25rem 1rem; border-radius: 9999px;
    background: linear-gradient(135deg, var(--primary), var(--accent));
    color: white; font-size: 0.75rem; font-weight: 600;
    white-space: nowrap;
  }
  .plan-header { text-align: center; margin-bottom: 1.5rem; }
  .plan-name {
    font-size: 1.125rem; font-weight: 700; color: var(--text);
    margin-bottom: 0.75rem;
  }
  .plan-price { display: flex; align-items: baseline; justify-content: center; gap: 0.25rem; }
  .price-value {
    font-size: 2.5rem; font-weight: 800; color: var(--text);
    letter-spacing: -0.03em;
  }
  .price-period { font-size: 0.875rem; color: var(--text-secondary); }
  .plan-features {
    list-style: none; display: flex; flex-direction: column; gap: 0.75rem;
    margin-bottom: 2rem; flex: 1;
  }
  .plan-features li {
    display: flex; align-items: center; gap: 0.625rem;
    font-size: 0.875rem; color: var(--text-secondary);
  }
  .plan-features svg { color: #10b981; flex-shrink: 0; }
  .plan-cta { width: 100%; padding: 0.875rem; font-weight: 600; }
  :global(.auth-spinner) {
    width: 18px; height: 18px; border: 2px solid rgba(255,255,255,0.3);
    border-top-color: white; border-radius: 50%;
    animation: spin 0.6s linear infinite; display: inline-block;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
  @media (max-width: 768px) {
    .pricing-grid { grid-template-columns: 1fr; max-width: 24rem; margin: 0 auto; }
    .pricing-card.highlighted { transform: none; }
    .pricing-card.highlighted:hover { transform: translateY(-4px); }
  }
</style>
