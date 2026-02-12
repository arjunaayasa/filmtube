import type { Config } from "tailwindcss"

const config: Config = {
  darkMode: "class",
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: "#ec1337",
        "background-light": "#f8f6f6",
        "background-dark": "#0f0f0f",
        surface: "#221013",
        "surface-dark": "#1e1e1e",
      },
      fontFamily: {
        display: ["var(--font-display)", "sans-serif"],
      },
      borderRadius: {
        DEFAULT: "0.25rem",
        lg: "0.5rem",
        xl: "0.75rem",
        full: "9999px",
      },
      gridTemplateColumns: {
        fluid: "repeat(auto-fill, minmax(300px, 1fr))",
      },
    },
  },
  plugins: [],
}

export default config
