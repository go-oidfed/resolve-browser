/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        'diff-added': '#22c55e',
        'diff-removed': '#ef4444',
        'diff-modified': '#eab308',
      },
    },
  },
  plugins: [],
}
