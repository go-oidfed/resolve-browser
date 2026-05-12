<script>
  import ResolveForm from './components/ResolveForm.svelte'
  import TrustChainView from './components/TrustChainView.svelte'
  import MetadataView from './components/MetadataView.svelte'
  import DiffView from './components/DiffView.svelte'
  import { result, error, showDiff, previewMetadata, editedChain, originalResult, trustAnchor, subject, entityTypes } from './lib/stores.js'
  import { previewEditedChain } from './lib/api.js'
  import { onMount } from 'svelte'
  
  let previewLoading = false
  let localError = null
  let previewError = null
  let urlCopied = false
  let darkMode = false
  
  onMount(() => {
    const saved = localStorage.getItem('darkMode')
    if (saved) {
      darkMode = saved === 'true'
    } else {
      darkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    
    if (darkMode) {
      document.documentElement.classList.add('dark')
    }
  })
  
  function toggleDarkMode() {
    darkMode = !darkMode
    localStorage.setItem('darkMode', darkMode.toString())
    
    if (darkMode) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }
  
  function copyURL() {
    navigator.clipboard.writeText(window.location.href)
    urlCopied = true
    setTimeout(() => urlCopied = false, 2000)
  }
  
  function handleEditChange(index, field, value) {
    $editedChain = $editedChain.map((stmt, i) => {
      if (i === index) {
        return { ...stmt, [field]: value }
      }
      return stmt
    })
  }
  
  async function previewResult() {
    previewLoading = true
    localError = null
    previewError = null
    
    try {
      const editedData = $editedChain.map((stmt, idx) => ({
        issuer: stmt.issuer,
        subject: stmt.subject,
        jwks: typeof stmt.jwks === 'string' ? JSON.parse(stmt.jwks) : stmt.jwks,
        metadata: typeof stmt.metadata === 'string' ? JSON.parse(stmt.metadata) : stmt.metadata,
        metadata_policy: typeof stmt.metadata_policy === 'string' ? JSON.parse(stmt.metadata_policy) : stmt.metadata_policy,
        constraints: typeof stmt.constraints === 'string' ? JSON.parse(stmt.constraints) : stmt.constraints,
        authority_hints: stmt.authority_hints,
        issued_at: stmt.iat,
        expires_at: stmt.exp
      }))
      
	const response = await previewEditedChain({
		trust_anchor: $result.trust_chain[0].issuer,
		trust_chain: [...editedData].reverse()
	})
      
      if (response.error) {
        previewError = {
          type: response.error,
          message: response.error_description || response.error
        }
        $previewMetadata = null
        $showDiff = false
      } else {
        $previewMetadata = response.metadata
        $showDiff = true
      }
    } catch (err) {
      previewError = {
        type: 'request_failed',
        message: err.message
      }
      localError = err.message
      $previewMetadata = null
      $showDiff = false
    } finally {
      previewLoading = false
    }
  }
  
  function resetToOriginal() {
    $editedChain = $originalResult.trust_chain.map(stmt => ({
      ...stmt,
      issuer: stmt.issuer,
      subject: stmt.subject,
      jwks: stmt.jwt_payload?.jwks || { keys: [] },
      metadata: stmt.jwt_payload?.metadata || {},
      metadata_policy: stmt.jwt_payload?.metadata_policy || null,
      constraints: stmt.jwt_payload?.constraints || null,
      authority_hints: stmt.jwt_payload?.authority_hints || [],
      iat: stmt.iat,
      exp: stmt.exp,
      chainLength: $result.trust_chain.length
    }))
    $previewMetadata = null
    $showDiff = false
    localError = null
  }
</script>

<main class="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors">
  <div class="container mx-auto px-4 py-8 max-w-7xl">
    <header class="mb-8">
      <div class="flex items-start justify-between">
        <div>
          <h1 class="text-3xl font-bold text-gray-900 dark:text-white">OpenID Federation Trust Chain Resolver</h1>
          <p class="text-gray-600 dark:text-gray-400 mt-2">Resolve and inspect OpenID federation trust chains</p>
        </div>
        <button
          on:click={toggleDarkMode}
          class="p-2 rounded-lg bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 transition-colors"
          title={darkMode ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {#if darkMode}
            <svg class="w-6 h-6 text-yellow-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"/>
            </svg>
          {:else}
            <svg class="w-6 h-6 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"/>
            </svg>
          {/if}
        </button>
      </div>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <div class="space-y-6">
        <ResolveForm />
        
        {#if $result}
          <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-4">
            <div class="flex items-start justify-between gap-4">
              <div class="flex-1">
                <h3 class="text-blue-800 dark:text-blue-300 font-semibold mb-1">Inline Editing Enabled</h3>
                <p class="text-blue-700 dark:text-blue-400 text-sm">Edit metadata, policy, and constraints directly in the trust chain cards below.</p>
              </div>
              <div class="flex items-center gap-2">
                <button
                  on:click={copyURL}
                  class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-sm font-medium flex items-center gap-1"
                  title="Copy URL with parameters"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                  {urlCopied ? 'Copied!' : 'Copy URL'}
                </button>
                <span class="text-gray-300 dark:text-gray-600">|</span>
                <button
                  on:click={resetToOriginal}
                  class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-sm font-medium"
                >
                  Reset Changes
                </button>
              </div>
            </div>
          </div>
          
          <TrustChainView 
            chain={$result.trust_chain} 
            editedChain={$editedChain}
            onEditChange={handleEditChange}
          />
        {/if}
      </div>

      <div class="space-y-6">
        {#if $error}
          <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <h3 class="text-red-800 dark:text-red-300 font-semibold mb-2">Error</h3>
            <p class="text-red-700 dark:text-red-400">{$error}</p>
          </div>
        {/if}
        
        {#if localError}
          <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <h3 class="text-red-800 dark:text-red-300 font-semibold mb-2">Error</h3>
            <p class="text-red-700 dark:text-red-400">{localError}</p>
          </div>
        {/if}

        {#if $result}
          <MetadataView 
            metadata={$result.metadata} 
            title="Resolved Metadata"
          />
          
          <div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-xl font-semibold text-gray-900 dark:text-white">Preview Changes</h2>
              <button
                on:click={previewResult}
                disabled={previewLoading}
                class="px-4 py-2 bg-green-600 text-white rounded-md font-medium hover:bg-green-700 disabled:opacity-50 transition-colors"
              >
                {previewLoading ? 'Previewing...' : 'Preview Result'}
              </button>
            </div>
            <p class="text-gray-600 dark:text-gray-400 text-sm mb-4">
              After editing fields in the trust chain cards, click "Preview Result" to see how your changes affect the resolved metadata.
            </p>
          </div>
          
          {#if $previewMetadata && $showDiff}
            <DiffView 
              original={$result.metadata} 
              modified={$previewMetadata}
            />
          {:else if previewError}
            <div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-6">
              <div class="flex items-start gap-3">
                <svg class="w-6 h-6 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 018 0z"/>
                </svg>
                <div class="flex-1">
                  <h3 class="text-red-800 dark:text-red-300 font-semibold mb-2">Preview Failed</h3>
                  <div class="text-sm text-red-700 dark:text-red-400">
                    <p class="font-mono bg-red-100 dark:bg-red-900/30 px-2 py-1 rounded mb-2 inline-block">{previewError.type}</p>
                    <p class="mt-2">{previewError.message}</p>
                  </div>
                  <p class="text-xs text-red-600 dark:text-red-400 mt-4 italic">
                    Fix the errors in your edited metadata and try again.
                  </p>
                </div>
              </div>
            </div>
          {/if}
        {/if}
      </div>
    </div>
  </div>
</main>
