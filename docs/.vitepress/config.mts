import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Dix",
  description: "Dix - Dependency Injection for Go",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: "Home", link: "/" },
      { text: "Documentation", link: "/docs/getting-started" },
      { text: "Github", link: "https://github.com/smtdfc/dix" },
    ],

    sidebar: {
      "/docs/": [
        {
          text: "Hướng dẫn bắt đầu",
          collapsed: false,
          items: [
            { text: "Bắt đầu", link: "/docs/getting-started" },
            { text: "Cài đặt", link: "/docs/installation" },
            { text: "Annotations", link: "/docs/annotations" },
            { text: "Singleton", link: "/docs/singleton" },
            { text: "Lỗi thường gặp", link: "/docs/common-errors" },
            { text: "Build and Run", link: "/docs/build" },
          ],
        },
      ],
    },
    footer: {
      message: "Released under the MIT License.",
      copyright: "Copyright © 2026-present smtdfc",
    },

    socialLinks: [{ icon: "github", link: "https://github.com/smtdfc/dix" }],
  },
});
