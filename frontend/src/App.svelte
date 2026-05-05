<script>
  import ResolveForm from './components/ResolveForm.svelte'
  import TrustChainView from './components/TrustChainView.svelte'
  import MetadataView from './components/MetadataView.svelte'
  import DiffView from './components/DiffView.svelte'
  import { result, error, showDiff, previewMetadata, editedChain, originalResult, trustAnchor, subject, entityTypes } from './lib/stores.js'
  import { previewEditedChain } from './lib/api.js'
  
  let previewLoading = false
  let localError = null
  let previewError = null
  let urlCopied = false
  
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
        trust_chain: editedData
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

<main class="min-h-screen bg-gray-50">
  <div class="container mx-auto px-4 py-8 max-w-7xl">
    <header class="mb-8">
      <h1 class="text-3xl font-bold text-gray-900">OpenID Federation Trust Chain Resolver</h1>
      <p class="text-gray-600 mt-2">Resolve and inspect OpenID federation trust chains</p>
    </header>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
      <div class="space-y-6">
        <ResolveForm />
        
        {#if $result}
          <div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
            <div class="flex items-start justify-between gap-4">
              <div class="flex-1">
                <h3 class="text-blue-800 font-semibold mb-1">Inline Editing Enabled</h3>
                <p class="text-blue-700 text-sm">Edit metadata, policy, and constraints directly in the trust chain cards below.</p>
              </div>
              <div class="flex items-center gap-2">
                <button
                  on:click={copyURL}
                  class="text-blue-600 hover:text-blue-800 text-sm font-medium flex items-center gap-1"
                  title="Copy URL with parameters"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                  {urlCopied ? 'Copied!' : 'Copy URL'}
                </button>
                <span class="text-gray-300">|</span>
                <button
                  on:click={resetToOriginal}
                  class="text-blue-600 hover:text-blue-800 text-sm font-medium"
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
          <div class="bg-red-50 border border-red-200 rounded-lg p-4">
            <h3 class="text-red-800 font-semibold mb-2">Error</h3>
            <p class="text-red-700">{$error}</p>
          </div>
        {/if}
        
        {#if localError}
          <div class="bg-red-50 border border-red-200 rounded-lg p-4">
            <h3 class="text-red-800 font-semibold mb-2">Error</h3>
            <p class="text-red-700">{localError}</p>
          </div>
        {/if}

        {#if $result}
          <MetadataView 
            metadata={$result.metadata} 
            title="Resolved Metadata"
          />
          
          <div class="bg-white rounded-lg shadow-md p-6">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-xl font-semibold text-gray-900">Preview Changes</h2>
              <button
                on:click={previewResult}
                disabled={previewLoading}
                class="px-4 py-2 bg-green-600 text-white rounded-md font-medium hover:bg-green-700 disabled:opacity-50 transition-colors"
              >
                {previewLoading ? 'Previewing...' : 'Preview Result'}
              </button>
            </div>
            <p class="text-gray-600 text-sm mb-4">
              After editing fields in the trust chain cards, click "Preview Result" to see how your changes affect the resolved metadata.
            </p>
          </div>
          
          {#if $previewMetadata && $showDiff}
            <DiffView 
              original={$result.metadata} 
              modified={$previewMetadata}
            />
          {:else if previewError}
            <div class="bg-red-50 border border-red-200 rounded-lg p-6">
              <div class="flex items-start gap-3">
                <svg class="w-6 h-6 text-red-600 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
                <div class="flex-1">
                  <h3 class="text-red-800 font-semibold mb-2">Preview Failed</h3>
                  <div class="text-sm text-red-700">
                    <p class="font-mono bg-red-100 px-2 py-1 rounded mb-2 inline-block">{previewError.type}</p>
                    <p class="mt-2">{previewError.message}</p>
                  </div>
                  <p class="text-xs text-red-600 mt-4 italic">
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
