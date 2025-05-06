/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        beige: '#f5f2ea',
        'pastel-brown': '#c8b6a6',
        'pastel-orange': '#f8b195',
        'pastel-green': '#a8d5ba',
        'pastel-purple': '#d0bdf4',
        'pastel-pink': '#f7d1cd',
        'pastel-red': '#ff6b6b'
      }
    },
  },
  plugins: [],
  safelist: [
    'bg-beige',
    'bg-pastel-brown',
    'bg-pastel-orange',
    'bg-pastel-green',
    'bg-pastel-purple',
    'bg-pastel-pink',
    'bg-pastel-red',
    'text-beige',
    'text-pastel-brown',
    'text-pastel-orange',
    'text-pastel-green',
    'text-pastel-purple',
    'text-pastel-pink',
    'text-pastel-red',
    'border-beige',
    'border-pastel-brown',
    'border-pastel-orange',
    'border-pastel-green',
    'border-pastel-purple',
    'border-pastel-pink',
    'border-pastel-red',
  ]
}