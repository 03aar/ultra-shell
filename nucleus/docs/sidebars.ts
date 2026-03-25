import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebars: SidebarsConfig = {
  docsSidebar: [
    "intro",
    "quickstart",
    {
      type: "category",
      label: "Installation",
      collapsed: false,
      items: [
        "installation/overview",
        "installation/macos",
        "installation/linux",
        "installation/windows",
        "installation/docker",
      ],
    },
    {
      type: "category",
      label: "Shell Runtime",
      collapsed: false,
      items: [
        "shell/overview",
        "shell/annotations",
        "shell/natural-language",
        "shell/configuration",
      ],
    },
    {
      type: "category",
      label: "MCP Integration",
      collapsed: false,
      items: [
        "mcp/overview",
        "mcp/claude-desktop",
        "mcp/tools-reference",
      ],
    },
    {
      type: "category",
      label: "REST API",
      collapsed: true,
      items: [
        "api/overview",
        "api/authentication",
        "api/executions",
        "api/context",
        "api/rollback",
        "api/sessions",
        "api/websocket",
      ],
    },
    {
      type: "category",
      label: "Agent Orchestration",
      collapsed: true,
      items: ["agents/overview", "agents/providers"],
    },
    {
      type: "category",
      label: "CLI",
      collapsed: true,
      items: ["cli/reference"],
    },
    {
      type: "category",
      label: "SDKs",
      collapsed: true,
      items: ["sdk/python", "sdk/typescript", "sdk/go"],
    },
    "contributing",
    "changelog",
  ],
};

export default sidebars;
