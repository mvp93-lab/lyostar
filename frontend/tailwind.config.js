/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        slate: {
          950: '#090a0f',
          900: '#11131b',
          850: '#161923',
          800: '#1e2230',
        },
        glacier: {
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
        }
      },
      borderColor: {
        subtle: 'rgba(255, 255, 255, 0.08)',
        subtleHover: 'rgba(255, 255, 255, 0.16)',
      }
    },
  },
  plugins: [],
}
