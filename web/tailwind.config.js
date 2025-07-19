/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Geist"', 'sans-serif'],
        mono: ['"Geist+Mono"', 'monospace'],
      },
    },
  },
  plugins: [require('@tailwindcss/typography')],
}
