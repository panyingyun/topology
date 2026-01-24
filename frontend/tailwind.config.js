/** @type {import('tailwindcss').Config} */
export default {
  darkMode: "class",
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        "primary": "#6366F1",
        "emerald-status": "#10B981",
        "background-light": "#f6f6f8",
        "background-dark": "#101122",
      },
      fontFamily: {
        "display": ["Inter", "sans-serif"]
      },
      borderRadius: {
        "DEFAULT": "0.5rem",
        "lg": "1rem",
        "xl": "1.5rem",
        "3xl": "1.5rem",
        "full": "9999px"
      },
      boxShadow: {
        'neu-flat': '20px 20px 60px #d1d1d6, -20px -20px 60px #ffffff',
        'neu-inset': 'inset 6px 6px 12px #d1d1d6, inset -6px -6px 12px #ffffff',
      }
    },
  },
  plugins: [],
}
