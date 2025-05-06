<script lang="ts">
  import '../app.css';
  import Navbar from '$lib/components/Navbar.svelte';
</script>

<div style="min-height: 100vh; background-color: #f5f2ea; display: flex; flex-direction: column;">
  <Navbar />

  <main style="flex: 1; max-width: 1200px; margin: 0 auto; padding: 2rem 1rem; width: 100%;">
    <slot />
  </main>

  <footer style="background-color: #c8b6a6; padding: 1rem; margin-top: auto;">
    <div style="max-width: 1200px; margin: 0 auto; text-align: center; color: white;">
      <p>© 2025 Throne of Thorns</p>
    </div>
  </footer>
</div>