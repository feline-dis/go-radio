// tailwind.config.js
import { fontFamily } from "tailwindcss/defaultTheme";

/** @type {import('tailwindcss').Config} */
const config = {
  darkMode: ["class"],
  content: [
    "./pages/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./app/**/*.{ts,tsx}",
    "./src/**/*.{ts,tsx}",
  ],
  theme: {
    container: {
      center: true,
      padding: "2rem",
      screens: {
        "2xl": "1400px",
      },
    },
    extend: {
      colors: {
        background: {
          DEFAULT: "hsl(0 0% 0%)",
          subtle: "hsl(0 0% 3%)",
          muted: "hsl(0 0% 7%)",
        },
        foreground: {
          DEFAULT: "hsl(0 0% 98%)",
          subtle: "hsl(0 0% 85%)",
          muted: "hsl(0 0% 65%)",
        },
        border: {
          DEFAULT: "hsl(0 0% 15%)",
          subtle: "hsl(0 0% 20%)",
        },
        primary: {
          DEFAULT: "hsl(0 0% 98%)",
          foreground: "hsl(0 0% 0%)",
        },
        accent: {
          DEFAULT: "hsl(0 0% 85%)",
          foreground: "hsl(0 0% 0%)",
        },
      },
      borderRadius: {
        lg: "0.5rem",
        md: "calc(0.5rem - 2px)",
        sm: "calc(0.5rem - 4px)",
      },
      fontFamily: {
        sans: ["var(--font-system)", ...fontFamily.sans],
        mono: ["var(--font-mono)", ...fontFamily.mono],
      },
      fontSize: {
        xs: ["0.75rem", { lineHeight: "1rem" }],
        sm: ["0.875rem", { lineHeight: "1.25rem" }],
        base: ["1rem", { lineHeight: "1.5rem" }],
        lg: ["1.125rem", { lineHeight: "1.75rem" }],
        xl: ["1.25rem", { lineHeight: "2rem" }],
      },
      boxShadow: {
        DEFAULT:
          "0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)",
      },
      animation: {
        progress: "progress 1s infinite linear",
      },
      keyframes: {
        progress: {
          "0%": { transform: "translateX(-100%)" },
          "100%": { transform: "translateX(100%)" },
        },
      },
    },
  },
  plugins: [],
};

export default config;
