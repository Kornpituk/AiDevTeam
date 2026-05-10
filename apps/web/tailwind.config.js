/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        status: {
          pending: "#6b7280",
          planning: "#3b82f6",
          approved: "#22c55e",
          in_progress: "#2563eb",
          reviewing: "#f97316",
          completed: "#16a34a",
          failed: "#dc2626",
        },
      },
    },
  },
  plugins: [],
}
