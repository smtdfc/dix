import { defineConfig } from "vitepress";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Dix",
  description: "Dix - Dependency Injection for Go",
  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: "Home", link: "/" },
      { text: "Github", link: "https://github.com/smtdfc/dix" },
    ],

    sidebar: [],
    footer: {
      message: "Released under the MIT License.",
      copyright: "Copyright © 2026-present smtdfc",
    },

    socialLinks: [{ icon: "github", link: "https://github.com/smtdfc/dix" }],
  },
});
