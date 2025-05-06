import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import tailwindcssVite from '@tailwindcss/vite';

export default defineConfig({
  plugins: [
    tailwindcssVite(),
    sveltekit()
  ]
});