<script lang="ts">
  import { auth, isAuthenticated } from '$lib/auth-store';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  onMount(() => {
    if (!$isAuthenticated) {
      goto('/login');
    }
  });
</script>

{#if $isAuthenticated}
  <div style="max-width: 48rem; margin: 0 auto;">
    <div style="background-color: white; padding: 1.5rem; border-radius: 0.5rem; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); margin-bottom: 2rem;">
      <h1 style="font-size: 1.75rem; font-weight: 700; margin-bottom: 1rem; color: #c8b6a6;">Welcome, {$auth.user?.username}!</h1>

      <div style="background-color: rgba(168, 213, 186, 0.2); padding: 1rem; border-radius: 0.375rem; margin-bottom: 1rem;">
        <p style="color: #a8d5ba; font-weight: 500;">You are successfully logged in! 🎉</p>
      </div>

      <p style="color: #666; line-height: 1.5;">
        This is the Throne of Thorns application. From here you can access all features
        of the platform. Use the navigation menu to explore different sections.
      </p>
    </div>

    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem;">
      <div style="background-color: white; padding: 1.5rem; border-radius: 0.5rem; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); display: flex; flex-direction: column; align-items: center; text-align: center;">
        <div style="width: 60px; height: 60px; border-radius: 50%; background-color: #f8b195; display: flex; justify-content: center; align-items: center; margin-bottom: 1rem; font-size: 1.5rem;">
          👑
        </div>
        <h3 style="font-size: 1.25rem; font-weight: 600; margin-bottom: 0.5rem; color: #f8b195;">Play Game</h3>
        <p style="color: #666; line-height: 1.5; margin-bottom: 1rem;">Start a new game session</p>
        <button style="padding: 0.5rem 1rem; background-color: #f8b195; color: white; border: none; border-radius: 0.375rem; font-weight: 500; cursor: pointer; transition: opacity 200ms; width: 100%;">
          Start Game
        </button>
      </div>

      <div style="background-color: white; padding: 1.5rem; border-radius: 0.5rem; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); display: flex; flex-direction: column; align-items: center; text-align: center;">
        <div style="width: 60px; height: 60px; border-radius: 50%; background-color: #d0bdf4; display: flex; justify-content: center; align-items: center; margin-bottom: 1rem; font-size: 1.5rem;">
          🏆
        </div>
        <h3 style="font-size: 1.25rem; font-weight: 600; margin-bottom: 0.5rem; color: #d0bdf4;">Leaderboard</h3>
        <p style="color: #666; line-height: 1.5; margin-bottom: 1rem;">See top players</p>
        <button style="padding: 0.5rem 1rem; background-color: #d0bdf4; color: white; border: none; border-radius: 0.375rem; font-weight: 500; cursor: pointer; transition: opacity 200ms; width: 100%;">
          View Scores
        </button>
      </div>

      <div style="background-color: white; padding: 1.5rem; border-radius: 0.5rem; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); display: flex; flex-direction: column; align-items: center; text-align: center;">
        <div style="width: 60px; height: 60px; border-radius: 50%; background-color: #a8d5ba; display: flex; justify-content: center; align-items: center; margin-bottom: 1rem; font-size: 1.5rem;">
          📊
        </div>
        <h3 style="font-size: 1.25rem; font-weight: 600; margin-bottom: 0.5rem; color: #a8d5ba;">Statistics</h3>
        <p style="color: #666; line-height: 1.5; margin-bottom: 1rem;">View your game stats</p>
        <button style="padding: 0.5rem 1rem; background-color: #a8d5ba; color: white; border: none; border-radius: 0.375rem; font-weight: 500; cursor: pointer; transition: opacity 200ms; width: 100%;">
          See Stats
        </button>
      </div>
    </div>
  </div>
{:else}
  <div style="text-align: center;">
    <p style="color: #ff6b6b;">Please log in first</p>
    <a href="/login" style="color: #d0bdf4; text-decoration: underline;">Go to Login</a>
  </div>
{/if}

<style>
  button:hover {
    opacity: 0.9;
  }
</style>