import { defineUserConfig } from 'vuepress'
import { defaultTheme } from '@vuepress/theme-default'
import { viteBundler } from '@vuepress/bundler-vite'
import { getSidebar } from './compose-sidenav'

const sidebar = {      
  '/guide/': [
    {
      text: 'Guide',
      children: [
        '/guide/overview/',
        '/guide/quick-start/',
        '/guide/architecture/',
        '/guide/adaptive-compression/',
        '/guide/auto-negotiation/',
        '/guide/external-tools/',
        '/guide/caching-concurrency/'
      ]
    }
  ],
  '/reference/': [
    {
      text: 'Reference',
      children: [
        '/reference/configuration/',
        '/reference/directives/',
        '/reference/errors/',
        '/reference/api/'
      ]
    }
  ],
  '/examples/': [
    {
      text: 'Examples',
      children: [
        '/examples/directives/',
        '/examples/nginx/'
      ]
    }
  ],
  '/contributing/': [
    {
      text: 'Contributing',
      children: [
        '/contributing/adding-format/',
        '/contributing/testing/'
      ]
    }
  ]
}

export default defineUserConfig({
  lang: 'en-US',
  title: "Piuma",
  description: "High-performance adaptive image optimization server in Go",
  base: '/piuma/',
  bundler: viteBundler(),
  theme: defaultTheme({
    navbar: [
      { text: 'Home', link: '/' },
      { text: 'Quick Start', link: '/guide/quick-start/' },
      { text: 'Guide', link: '/guide/' },
      { text: 'Reference', link: '/reference/' },
      { text: 'Examples', link: '/examples/' },
  { text: 'Contributing', link: '/contributing/' },
  { text: 'Changelog', link: '/changelog/' }
    ],
    // sidebar: getSidebar(),
    sidebar: sidebar,
    repo: 'piumaio/piuma',
    smoothScroll: true,
    logo: '/images/logo.png',
    editLink: false
  }),
  head: [
    ['link', { rel: "icon", href: "data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>🪶</text></svg>" }],
    ['link', { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css?family=Roboto:300,400,500,700|Material+Icons' }],
    ['link', { rel: 'stylesheet', href: 'https://cdn.jsdelivr.net/npm/@mdi/font@4.x/css/materialdesignicons.min.css' }],
  ],
})
