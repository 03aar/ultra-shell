import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";

const config: Config = {
  title: "Nucleus",
  tagline: "The AI-Native Shell Runtime",
  favicon: "img/favicon.ico",

  url: "https://docs.nucleusshell.dev",
  baseUrl: "/",

  organizationName: "03aar",
  projectName: "ultra-shell",

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      {
        docs: {
          sidebarPath: "./sidebars.ts",
          editUrl:
            "https://github.com/03aar/ultra-shell/tree/main/nucleus/docs/",
        },
        blog: false,
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      defaultMode: "dark",
      disableSwitch: true,
      respectPrefersColorScheme: false,
    },
    navbar: {
      title: "Nucleus",
      logo: {
        alt: "Nucleus Logo",
        src: "img/logo.svg",
      },
      items: [
        {
          type: "docSidebar",
          sidebarId: "docsSidebar",
          position: "left",
          label: "Docs",
        },
        {
          to: "/docs/api/overview",
          label: "API",
          position: "left",
        },
        {
          href: "https://github.com/03aar/ultra-shell",
          label: "GitHub",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Product",
          items: [
            { label: "Getting Started", to: "/docs/intro" },
            { label: "Quickstart", to: "/docs/quickstart" },
            { label: "CLI Reference", to: "/docs/cli/reference" },
            { label: "API Reference", to: "/docs/api/overview" },
          ],
        },
        {
          title: "Community",
          items: [
            {
              label: "GitHub",
              href: "https://github.com/03aar/ultra-shell",
            },
            {
              label: "Discussions",
              href: "https://github.com/03aar/ultra-shell/discussions",
            },
            {
              label: "Issues",
              href: "https://github.com/03aar/ultra-shell/issues",
            },
          ],
        },
        {
          title: "More",
          items: [
            { label: "Changelog", to: "/docs/changelog" },
            { label: "Contributing", to: "/docs/contributing" },
            {
              label: "License (MIT)",
              href: "https://github.com/03aar/ultra-shell/blob/main/LICENSE",
            },
          ],
        },
      ],
      copyright: `Copyright \u00a9 ${new Date().getFullYear()} Nucleus Contributors. Released under the MIT License.`,
    },
    prism: {
      theme: prismThemes.vsDark,
      darkTheme: prismThemes.vsDark,
      additionalLanguages: [
        "bash",
        "json",
        "yaml",
        "toml",
        "rust",
        "go",
        "python",
        "typescript",
        "docker",
      ],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
