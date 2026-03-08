---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "Dix"
  text: "Zero-Magic Dependency Injection"
  tagline: "Explicit, Compile-time safe, and Reflection-free DI generator for Go."
  actions:
    - theme: brand
      text: Get Started
      link: /docs/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/smtdfc/dix
features:
  - title: "Zero Reflection"
    details: "No runtime overhead. Dix generates standard Go code that initializes your dependencies at compile-time."
    icon: ⚡
  - title: "Compile-time Safety"
    details: "Catch missing dependencies or circular loops during development, not in production. If it builds, it works."
    icon: 🛡️
  - title: "No Magic"
    details: "Generated code is clean and human-readable. No hidden behavior, no mysterious package global states."
    icon: 🔍
  - title: "Explicit over Implicit"
    details: "Strict signature matching ensures your dependency graph is always predictable and easy to debug."
    icon: 📐

footer:
  message: "Made with ❤️ by smtdfc"
  copyright: "Copyright © 2026 Dix Project"
  links:
    - title: "Community"
      items:
        - text: "GitHub"
          link: "https://github.com/smtdfc/dix"
        - text: "Issues"
          link: "https://github.com/smtdfc/dix/issues"
    - title: "Ecosystem"
      items:
        - text: "Go Official"
          link: "https://go.dev"
        - text: "Cobra CLI"
          link: "https://github.com/spf13/cobra"
---
