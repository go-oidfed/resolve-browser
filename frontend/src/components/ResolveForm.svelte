<script>
  import { trustAnchor, subject, entityTypes, loading, error, result, originalResult, editedChain, showDiff, previewMetadata, updateURL, loadFromURL } from '../lib/stores.js'
  import { resolveTrustChain } from '../lib/api.js'
  import { onMount } from 'svelte'
  
  const availableEntityTypes = [
    'federation_entity',
    'openid_provider',
    'openid_relying_party',
    'oauth_authorization_server',
    'oauth_resource_server'
  ]
  
  onMount(() => {
    const params = loadFromURL()
    if (params.trustAnchor) $trustAnchor = params.trustAnchor
    if (params.subject) $subject = params.subject
    if (params.entityTypes.length > 0) $entityTypes = params.entityTypes
  })
  
  async function handleSubmit() {
    $loading = true
    $error = null
    
    try {
      const data = {
        sub: $subject,
        trust_anchor: [$trustAnchor],
        entity_types: $entityTypes.length > 0 ? $entityTypes : undefined
      }
      
      const response = await resolveTrustChain(data)
      $result = response
      $originalResult = JSON.parse(JSON.stringify(response))
      $editedChain = response.trust_chain.map((stmt, index) => ({
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
        chainLength: response.trust_chain.length
      }))
      $previewMetadata = null
      $showDiff = false
      
      updateURL($trustAnchor, $subject, $entityTypes)
    } catch (err) {
      $error = err.message
      $result = null
    } finally {
      $loading = false
    }
  }
  
  function handleClear() {
    $trustAnchor = ''
    $subject = ''
    $entityTypes = []
    $result = null
    $originalResult = null
    $editedChain = []
    $error = null
    $previewMetadata = null
    $showDiff = false
  }
  
  function toggleEntityType(type) {
    if ($entityTypes.includes(type)) {
      $entityTypes = $entityTypes.filter(t => t !== type)
    } else {
      $entityTypes = [...$entityTypes, type]
    }
  }
</script>

<div class="bg-white dark:bg-gray-800 rounded-lg shadow-md p-6">
  <h2 class="text-xl font-semibold text-gray-900 dark:text-white mb-4">Resolve Trust Chain</h2>
  
  <div class="space-y-4">
    <div>
      <label for="trust-anchor" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
        Trust Anchor URL *
      </label>
      <input
        id="trust-anchor"
        type="url"
        bind:value={$trustAnchor}
        placeholder="https://ta.example.com"
        class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        required
      />
    </div>
    
    <div>
      <label for="subject" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
        Subject Entity ID *
      </label>
      <input
        id="subject"
        type="url"
        bind:value={$subject}
        on:keydown={(e) => {
          if (e.key === 'Enter') {
            handleSubmit()
          }
        }}
        placeholder="https://rp.example.com"
        class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        required
      />
    </div>
    
    <div>
      <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        Entity Types (optional)
      </label>
      <div class="flex flex-wrap gap-2">
        {#each availableEntityTypes as type}
          <button
            type="button"
            on:click={() => toggleEntityType(type)}
            class="px-3 py-1 text-sm rounded-full transition-colors {
              $entityTypes.includes(type)
                ? 'bg-blue-500 text-white'
                : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
            }"
          >
            {type}
          </button>
        {/each}
      </div>
    </div>
    
    <div class="flex gap-3 pt-4">
      <button
        on:click={handleSubmit}
        disabled={$loading || !$trustAnchor || !$subject}
        class="flex-1 bg-blue-600 text-white px-4 py-2 rounded-md font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {#if $loading}
          <span class="flex items-center justify-center gap-2">
            <svg class="animate-spin h-5 w-5" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none"/>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
            </svg>
            Resolving...
          </span>
        {:else}
          Resolve
        {/if}
      </button>
      
      <button
        on:click={handleClear}
        class="px-4 py-2 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-md font-medium hover:bg-gray-50 dark:hover:bg-gray-700 focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 dark:focus:ring-offset-gray-800 transition-colors"
      >
        Clear
      </button>
    </div>
  </div>
</div>
