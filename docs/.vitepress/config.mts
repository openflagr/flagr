import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

const SITE_URL = 'https://openflagr.github.io'
const BASE = '/flagr/'
const CANONICAL_ORIGIN = `${SITE_URL}${BASE}`

const description =
  'Flagr — open-source Go service for feature flags, A/B testing, and dynamic configuration. Self-hosted, API-first, GitOps-friendly.'

const organizationJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'Organization',
  name: 'openflagr',
  url: 'https://github.com/openflagr',
  logo: `${CANONICAL_ORIGIN}images/logo.svg`,
  sameAs: ['https://github.com/openflagr/flagr'],
}

const websiteJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'WebSite',
  name: 'Flagr Docs',
  url: CANONICAL_ORIGIN,
  description,
  publisher: {
    '@type': 'Organization',
    name: 'openflagr',
    url: 'https://github.com/openflagr',
  },
}

const softwareJsonLd = {
  '@context': 'https://schema.org',
  '@type': 'SoftwareApplication',
  name: 'Flagr',
  applicationCategory: 'DeveloperApplication',
  operatingSystem: 'Linux, macOS, Windows',
  description,
  url: CANONICAL_ORIGIN,
  downloadUrl: 'https://github.com/openflagr/flagr',
  installUrl: 'https://github.com/openflagr/flagr/pkgs/container/flagr',
  softwareVersion: 'latest',
  license: 'https://github.com/openflagr/flagr/blob/main/LICENSE',
  offers: {
    '@type': 'Offer',
    price: '0',
    priceCurrency: 'USD',
  },
  author: {
    '@type': 'Organization',
    name: 'openflagr',
    url: 'https://github.com/openflagr',
  },
  codeRepository: 'https://github.com/openflagr/flagr',
  programmingLanguage: 'Go',
}

function jsonLdScript(data: Record<string, unknown>) {
  return [
    'script',
    { type: 'application/ld+json' },
    JSON.stringify(data),
  ] as [string, Record<string, string>, string]
}

export default withMermaid(
  defineConfig({
    title: 'Flagr',
    description,
    base: BASE,
    lang: 'en-US',
    cleanUrls: true,
    // Stock VitePress default theme, light only (no dark toggle).
    appearance: false,
    // Fail the build on broken internal links (plans/ are GitHub URLs, not site routes).
    ignoreDeadLinks: false,
    srcExclude: ['**/plans/**', '**/api_docs/**', '**/theme/**', '**/snippets/**'],

    head: [
      ['link', { rel: 'icon', href: `${BASE}favicon.png`, type: 'image/png' }],
      [
        'link',
        {
          rel: 'apple-touch-icon',
          href: `${BASE}apple-touch-icon.png`,
        },
      ],
      ['meta', { name: 'theme-color', content: '#3451b2' }],
      ['meta', { name: 'author', content: 'openflagr' }],
      [
        'meta',
        {
          name: 'keywords',
          content:
            'feature flags, feature toggles, A/B testing, dynamic configuration, open source, Go, self-hosted, GitOps, Flagr',
        },
      ],
      ['meta', { property: 'og:type', content: 'website' }],
      ['meta', { property: 'og:site_name', content: 'Flagr Docs' }],
      ['meta', { property: 'og:title', content: 'Flagr Docs' }],
      ['meta', { property: 'og:description', content: description }],
      ['meta', { property: 'og:url', content: CANONICAL_ORIGIN }],
      [
        'meta',
        {
          property: 'og:image',
          content: `${CANONICAL_ORIGIN}images/demo_homepage.png`,
        },
      ],
      ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
      ['meta', { name: 'twitter:title', content: 'Flagr Docs' }],
      ['meta', { name: 'twitter:description', content: description }],
      [
        'meta',
        {
          name: 'twitter:image',
          content: `${CANONICAL_ORIGIN}images/demo_homepage.png`,
        },
      ],
      jsonLdScript(organizationJsonLd),
      jsonLdScript(websiteJsonLd),
      jsonLdScript(softwareJsonLd),
      // Docsify hash-route → clean path (old bookmarks / external links).
      // Maps #/page?id=section → /flagr/page#section
      [
        'script',
        {},
        `(function(){var h=location.hash;if(!h||h.charAt(1)!=='/')return;var raw=h.slice(2).replace(/\\.md$/i,'');var id='';var qi=raw.indexOf('?');if(qi>=0){var qs=raw.slice(qi+1);raw=raw.slice(0,qi);var m=qs.match(/(?:^|&)id=([^&]+)/);if(m)id=decodeURIComponent(m[1]);}var p=raw.replace(/\\/+$/,'');var base='${BASE}'.replace(/\\/$/,'');var url;if(!p||p==='home'||p==='index'){url=base+'/';}else{url=base+'/'+p;}if(id)url+='#'+id;location.replace(url);})();`,
      ],
    ],

    themeConfig: {
      // Option D lockup (SVG monogram + FLAGR); hide duplicate site title text
      logo: {
        light: '/images/logo.svg',
        dark: '/images/logo.svg',
        alt: 'Flagr',
      },
      siteTitle: false,
      nav: [
        { text: 'Get started', link: '/' },
        { text: 'Use cases', link: '/flagr_use_cases' },
        { text: 'Self-hosting', link: '/flagr_self_host' },
        {
          text: 'API reference',
          link: `${CANONICAL_ORIGIN}api_docs/`,
          target: '_blank',
          rel: 'noopener noreferrer',
        },
        {
          text: 'GitHub',
          link: 'https://github.com/openflagr/flagr',
          target: '_blank',
          rel: 'noopener noreferrer',
        },
      ],
      sidebar: [
        {
          text: 'Documentation',
          items: [
            { text: 'Home', link: '/' },
            { text: 'Integration guide', link: '/integration' },
            {
              text: 'Behavioral contracts',
              link: '/flagr_behavioral_contracts',
            },
            { text: 'Contributing', link: '/CONTRIBUTING' },
          ],
        },
        {
          text: 'Concepts & API',
          items: [
            { text: 'Overview', link: '/flagr_overview' },
            { text: 'Use cases', link: '/flagr_use_cases' },
            {
              text: 'Built-in context injection',
              link: '/flagr_injected_context',
            },
            { text: 'Debug console', link: '/flagr_debugging' },
            {
              text: 'API reference',
              link: `${CANONICAL_ORIGIN}api_docs/`,
              target: '_blank',
              rel: 'noopener noreferrer',
            },
          ],
        },
        {
          text: 'Deploy & config',
          items: [
            { text: 'Self-hosting', link: '/flagr_self_host' },
            { text: 'Environment variables', link: '/flagr_env' },
            {
              text: 'JSON flag source (GitOps)',
              link: '/flagr_json_flag_spec',
            },
            { text: 'Notifications', link: '/flagr_notifications' },
          ],
        },
        {
          text: 'Analytics',
          items: [
            { text: 'Exposure logging', link: '/flagr_exposure' },
            {
              text: 'Data recorders & A/B analysis',
              link: '/flagr_eval_exposure_pipeline',
            },
            { text: 'Datar analytics', link: '/flagr_datar' },
          ],
        },
        {
          text: 'Development',
          items: [{ text: 'Testing', link: '/flagr_testing' }],
        },
      ],
      socialLinks: [
        { icon: 'github', link: 'https://github.com/openflagr/flagr' },
      ],
      search: {
        provider: 'local',
      },
      editLink: {
        pattern: 'https://github.com/openflagr/flagr/edit/main/docs/:path',
        text: 'Edit this page on GitHub',
      },
      outline: {
        level: [2, 3],
      },
      footer: {
        message: 'Open source under the Apache-2.0 License.',
        copyright: 'Copyright © openflagr contributors',
      },
      externalLinkIcon: true,
    },

    markdown: {
      theme: {
        light: 'github-light',
        dark: 'github-dark',
      },
      languages: ['bash', 'go', 'js', 'json', 'yaml', 'shell', 'sh'],
    },

    sitemap: {
      hostname: CANONICAL_ORIGIN,
    },

    transformPageData(pageData) {
      const path =
        pageData.relativePath === 'index.md'
          ? ''
          : pageData.relativePath.replace(/\.md$/, '')
      const pageUrl = `${CANONICAL_ORIGIN}${path}`
      const title = pageData.title
        ? `${pageData.title} | Flagr`
        : 'Flagr Docs'
      const desc = pageData.description || description

      pageData.frontmatter.head ??= []
      pageData.frontmatter.head.push(
        ['link', { rel: 'canonical', href: pageUrl }],
        ['meta', { property: 'og:title', content: title }],
        ['meta', { property: 'og:description', content: desc }],
        ['meta', { property: 'og:url', content: pageUrl }],
        ['meta', { name: 'twitter:title', content: title }],
        ['meta', { name: 'twitter:description', content: desc }],
      )
    },
  }),
)
