import React from "react";
import clsx from "clsx";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import styles from "./index.module.css";

function HeroSection() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={styles.hero}>
      <div className="container">
        <h1 className={styles.heroTitle}>{siteConfig.title}</h1>
        <p className={styles.heroTagline}>{siteConfig.tagline}</p>
        <p className={styles.heroDescription}>
          A drop-in shell replacement that intercepts every command, evaluates
          risk, tracks mutations, enables rollback, and exposes everything to AI
          agents through MCP and REST APIs.
        </p>
        <div className={styles.heroActions}>
          <Link className={styles.primaryButton} to="/docs/quickstart">
            Get Started
          </Link>
          <Link className={styles.secondaryButton} to="/docs/intro">
            Learn More
          </Link>
        </div>
        <div className={styles.installBlock}>
          <code>curl -fsSL https://nucleusshell.dev/install.sh | sh</code>
        </div>
      </div>
    </header>
  );
}

type FeatureItem = {
  title: string;
  description: string;
  icon: string;
};

const features: FeatureItem[] = [
  {
    title: "Shell Runtime",
    description:
      "Wraps your existing shell via PTY interception. Every command is parsed, risk-evaluated, and recorded as a node in an execution DAG. File mutations are snapshotted for rollback.",
    icon: ">_",
  },
  {
    title: "Execution Rollback",
    description:
      "Made a mistake? Nucleus snapshots files before modification. Roll back any command instantly via CLI, API, or natural language through an AI agent.",
    icon: "<-",
  },
  {
    title: "MCP Server",
    description:
      "Fully compliant Model Context Protocol server. Connect Claude Desktop, Cursor, or Zed and let AI agents execute commands, read context, and manage your shell.",
    icon: "{}",
  },
  {
    title: "Agent Orchestration",
    description:
      "Give Claude, GPT-4, Gemini, or Ollama structured access to your shell. Agents plan multi-step workflows, execute with guardrails, and observe results.",
    icon: "AI",
  },
  {
    title: "Risk Evaluation",
    description:
      "Every command is evaluated for risk before execution. Destructive operations are flagged or blocked. Dry-run mode lets you preview impact without executing.",
    icon: "!!",
  },
  {
    title: "Real-Time Dashboard",
    description:
      "Watch commands execute in real time. Visualize the execution DAG, replay sessions, explore API endpoints, and manage agent workflows from the browser.",
    icon: "[]",
  },
];

function FeatureCard({ title, description, icon }: FeatureItem) {
  return (
    <div className={styles.featureCard}>
      <div className={styles.featureIcon}>{icon}</div>
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  );
}

function FeaturesSection() {
  return (
    <section className={styles.features}>
      <div className="container">
        <h2 className={styles.sectionTitle}>What Nucleus Does</h2>
        <div className={styles.featureGrid}>
          {features.map((props, idx) => (
            <FeatureCard key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}

function ArchitectureSection() {
  return (
    <section className={styles.architecture}>
      <div className="container">
        <h2 className={styles.sectionTitle}>Three Layers, One Platform</h2>
        <div className={styles.layerGrid}>
          <div className={styles.layer}>
            <span className={styles.layerTag}>Layer 1</span>
            <h3>nucleus-core</h3>
            <p className={styles.layerLang}>Rust</p>
            <p>
              PTY interception, command parsing, risk evaluation, file
              snapshots, execution DAG, sled storage. The foundation everything
              else builds on.
            </p>
          </div>
          <div className={styles.layer}>
            <span className={styles.layerTag}>Layer 2</span>
            <h3>nucleus-api</h3>
            <p className={styles.layerLang}>Go</p>
            <p>
              REST and WebSocket API server. Executions, context, sessions,
              rollback, agent commands, skills. Connects to PostgreSQL and
              Redis.
            </p>
          </div>
          <div className={styles.layer}>
            <span className={styles.layerTag}>Layer 3</span>
            <h3>nucleus-dashboard</h3>
            <p className={styles.layerLang}>Next.js</p>
            <p>
              Real-time execution feed, DAG visualization with ReactFlow,
              terminal replay with xterm.js, agent console, and API explorer.
            </p>
          </div>
        </div>
      </div>
    </section>
  );
}

export default function Home(): React.JSX.Element {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout title="Home" description={siteConfig.tagline}>
      <HeroSection />
      <main>
        <FeaturesSection />
        <ArchitectureSection />
      </main>
    </Layout>
  );
}
