<script lang="ts">
  import { auth, isAuthenticated } from '$lib/auth-store';
  import { goto } from '$app/navigation';

  function handleLogout() {
    auth.clearAuth();
    goto('/login');
  }
</script>

<nav style="background-color: #c8b6a6; padding: 0.75rem 1rem;">
  <div style="display: flex; justify-content: space-between; align-items: center; max-width: 1200px; margin: 0 auto;">
    <div>
      <a href="/" style="color: white; font-size: 1.5rem; font-weight: 700; text-decoration: none;">
        Throne of Thorns
      </a>
    </div>

    {#if $isAuthenticated}
      <div style="display: flex; gap: 1.5rem;">
        <a
          href="/"
          style="color: white; text-decoration: none; font-weight: 500; transition: opacity 200ms;"
          class="nav-link"
        >
          Home
        </a>
        <a
          href="/profile"
          style="color: white; text-decoration: none; font-weight: 500; transition: opacity 200ms;"
          class="nav-link"
        >
          Profile
        </a>
        <button
          on:click={handleLogout}
          style="background: none; border: none; color: white; font-weight: 500; cursor: pointer; padding: 0; transition: opacity 200ms;"
          class="nav-link"
        >
          Logout
        </button>
      </div>
    {:else}
      <div style="display: flex; gap: 1.5rem;">
        <a
          href="/login"
          style="color: white; text-decoration: none; font-weight: 500; transition: opacity 200ms;"
          class="nav-link"
        >
          Login
        </a>
        <a
          href="/signup"
          style="color: white; text-decoration: none; font-weight: 500; transition: opacity 200ms;"
          class="nav-link"
        >
          Sign Up
        </a>
      </div>
    {/if}
  </div>
</nav>

<style>
  .nav-link:hover {
    opacity: 0.8;
    text-decoration: none !important;
  }
</style>