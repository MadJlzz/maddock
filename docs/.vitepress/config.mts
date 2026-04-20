import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Maddock',
  description: 'An infrastructure-as-code tool for Linux',
  cleanUrls: true,
  base: '/maddock/',

  themeConfig: {
    nav: [
      { text: 'Guide', link: '/installation' },
      { text: 'Reference', link: '/cli/agent' },
    ],

    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Overview', link: '/' },
          { text: 'Installation', link: '/installation' },
          { text: 'Architecture', link: '/architecture' },
        ],
      },
      {
        text: 'CLI Reference',
        items: [
          { text: 'maddock-agent', link: '/cli/agent' },
          { text: 'maddock-server', link: '/cli/server' },
        ],
      },
      {
        text: 'Resources',
        items: [
          { text: 'Overview', link: '/resources/' },
          { text: 'package', link: '/resources/package' },
          { text: 'file', link: '/resources/file' },
          { text: 'service', link: '/resources/service' },
        ],
      },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/MadJlzz/maddock' },
    ],

    editLink: {
      pattern: 'https://github.com/MadJlzz/maddock/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },

    search: {
      provider: 'local',
    },

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2026 Julien Klaer',
    },
  },
})
