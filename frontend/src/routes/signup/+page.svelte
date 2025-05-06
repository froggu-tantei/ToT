<script lang="ts">
  import { signUp } from '$lib/api';
  import { auth, isAuthenticated } from '$lib/auth-store';
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  let email = '';
  let username = '';
  let password = '';
  let confirmPassword = '';
  let error = '';
  let loading = false;

  onMount(() => {
    if ($isAuthenticated) {
      goto('/');
    }
  });

  async function handleSubmit() {
    // Reset error
    error = '';

    // Form validation
    if (!email || !username || !password) {
      error = 'All fields are required';
      return;
    }

    if (password !== confirmPassword) {
      error = 'Passwords do not match';
      return;
    }

    loading = true;
    try {
      const result = await signUp(email, username, password);

      if (result.success && result.data) {
        // Set auth data and redirect to home
        auth.setAuth(result.data.user, result.data.token);
        goto('/');
      } else {
        error = result.error || 'Something went wrong with signup';
      }
    } catch (err) {
      console.error('Signup error:', err);
      error = 'An unexpected error occurred';
    } finally {
      loading = false;
    }
  }
</script>

<div style="max-width: 28rem; margin: 0 auto; background-color: white; padding: 1.5rem; border-radius: 0.5rem; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);">
  <h1 style="font-size: 1.5rem; font-weight: 700; text-align: center; margin-bottom: 1.5rem; color: #c8b6a6;">Sign Up</h1>

  {#if error}
    <div style="background-color: rgba(255, 107, 107, 0.2); padding: 0.75rem; border-radius: 0.375rem; margin-bottom: 1rem;">
      <p style="color: #ff6b6b;">{error}</p>
    </div>
  {/if}

  <form on:submit|preventDefault={handleSubmit}>
    <div style="margin-bottom: 1rem;">
      <label for="email" style="display: block; margin-bottom: 0.25rem; font-weight: 500; color: #c8b6a6;">Email</label>
      <input
        type="email"
        id="email"
        bind:value={email}
        style="width: 100%; padding: 0.5rem 0.75rem; border: 1px solid #c8b6a6; border-radius: 0.375rem;"
        placeholder="your@email.com"
        required
      />
    </div>

    <div style="margin-bottom: 1rem;">
      <label for="username" style="display: block; margin-bottom: 0.25rem; font-weight: 500; color: #c8b6a6;">Username</label>
      <input
        type="text"
        id="username"
        bind:value={username}
        style="width: 100%; padding: 0.5rem 0.75rem; border: 1px solid #c8b6a6; border-radius: 0.375rem;"
        placeholder="cooluser123"
        required
      />
    </div>

    <div style="margin-bottom: 1rem;">
      <label for="password" style="display: block; margin-bottom: 0.25rem; font-weight: 500; color: #c8b6a6;">Password</label>
      <input
        type="password"
        id="password"
        bind:value={password}
        style="width: 100%; padding: 0.5rem 0.75rem; border: 1px solid #c8b6a6; border-radius: 0.375rem;"
        placeholder="••••••••"
        required
      />
    </div>

    <div style="margin-bottom: 1.5rem;">
      <label for="confirmPassword" style="display: block; margin-bottom: 0.25rem; font-weight: 500; color: #c8b6a6;">Confirm Password</label>
      <input
        type="password"
        id="confirmPassword"
        bind:value={confirmPassword}
        style="width: 100%; padding: 0.5rem 0.75rem; border: 1px solid #c8b6a6; border-radius: 0.375rem;"
        placeholder="••••••••"
        required
      />
    </div>

    <button
      type="submit"
      style="width: 100%; padding: 0.5rem 1rem; background-color: #d0bdf4; color: white; border-radius: 0.375rem; font-weight: 500; transition: opacity 200ms;"
      disabled={loading}
    >
      {loading ? 'Creating account...' : 'Sign Up'}
    </button>

    <div style="margin-top: 1rem; text-align: center;">
      <p style="color: #c8b6a6;">
        Already have an account?
        <a href="/login" style="color: #d0bdf4; text-decoration: none;">Log in</a>
      </p>
    </div>
  </form>
</div>