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
  <div style="max-width: 32rem; margin: 0 auto; background-color: white; padding: 1.5rem; border-radius: 0.5rem; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);">
    <div style="text-align: center; margin-bottom: 1.5rem;">
      <div style="width: 100px; height: 100px; border-radius: 50%; background-color: #d0bdf4; display: flex; justify-content: center; align-items: center; margin: 0 auto; color: white; font-size: 2rem; font-weight: bold;">
        {$auth.user?.username.slice(0, 1).toUpperCase()}
      </div>
      <h1 style="font-size: 1.5rem; font-weight: 700; margin-top: 1rem; color: #c8b6a6;">{$auth.user?.username}</h1>
    </div>

    <div style="background-color: #f9f7f3; padding: 1.25rem; border-radius: 0.5rem; margin-bottom: 1.5rem;">
      <h2 style="font-size: 1.25rem; font-weight: 600; margin-bottom: 1rem; color: #d0bdf4; border-bottom: 1px solid #ebe6d9; padding-bottom: 0.5rem;">Account Information</h2>

      <div style="margin-bottom: 0.75rem;">
        <p style="margin-bottom: 0.25rem; color: #888;">Email Address</p>
        <p style="font-weight: 500;">{$auth.user?.email}</p>
      </div>

      <div style="margin-bottom: 0.75rem;">
        <p style="margin-bottom: 0.25rem; color: #888;">Username</p>
        <p style="font-weight: 500;">{$auth.user?.username}</p>
      </div>

      <div>
        <p style="margin-bottom: 0.25rem; color: #888;">Member Since</p>
        <p style="font-weight: 500;">{new Date($auth.user?.created_at || '').toLocaleDateString()}</p>
      </div>
    </div>

    <div style="display: flex; flex-direction: column; gap: 0.75rem;">
      <button
        style="padding: 0.75rem; background-color: #f8b195; color: white; border: none; border-radius: 0.375rem; font-weight: 500; cursor: pointer; transition: opacity 200ms;"
      >
        Edit Profile
      </button>

      <button
        style="padding: 0.75rem; background-color: #a8d5ba; color: white; border: none; border-radius: 0.375rem; font-weight: 500; cursor: pointer; transition: opacity 200ms;"
      >
        Change Password
      </button>
    </div>
  </div>
{:else}
  <div style="text-align: center;">
    <p style="color: #ff6b6b;">Please log in to view your profile</p>
    <a href="/login" style="color: #d0bdf4; text-decoration: underline;">Go to Login</a>
  </div>
{/if}

<style>
  button:hover {
    opacity: 0.9;
  }
</style>