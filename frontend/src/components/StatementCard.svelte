<script>
  import JWTViewer from './JWTViewer.svelte'
  import { formatTime, formatEntityId } from '../lib/utils.js'
  
  export let statement
  export let index
  export let isExpanded = false
  export let onToggle
  export let editedStatement
  export let onEditChange
  
  function getTypeLabel(type) {
    const labels = {
      'trust_anchor_entity_configuration': 'Trust Anchor Entity Configuration',
      'subordinate_statement': 'Subordinate Statement',
      'entity_configuration': 'Entity Configuration'
    }
    return labels[type] || type
  }
  
  function getTypeColor(type) {
    const colors = {
      'trust_anchor_entity_configuration': 'bg-green-100 dark:bg-green-900/30 border-green-300 dark:border-green-700 text-green-800 dark:text-green-300',
      'subordinate_statement': 'bg-blue-100 dark:bg-blue-900/30 border-blue-300 dark:border-blue-700 text-blue-800 dark:text-blue-300',
      'entity_configuration': 'bg-purple-100 dark:bg-purple-900/30 border-purple-300 dark:border-purple-700 text-purple-800 dark:text-purple-300'
    }
    return colors[type] || 'bg-gray-100 dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-800 dark:text-gray-300'
  }
  
  function isTrustAnchor() {
    return index === 0
  }
  
  function isLeafStatement() {
    return index === editedStatement?.chainLength - 1
  }
  
  function isDirectSubordinate() {
    return index === editedStatement?.chainLength - 2
  }
  
  function canEditMetadata() {
    // Trust Anchor cannot be edited, only leaf and direct subordinate
    return !isTrustAnchor() && (isLeafStatement() || isDirectSubordinate())
  }
  
  function canEditMetadataPolicy() {
    // Trust Anchor cannot be edited, all others except leaf can have policy
    return !isTrustAnchor() && !isLeafStatement()
  }
  
  function canEditConstraints() {
    // Trust Anchor cannot be edited, all others except leaf can have constraints
    return !isTrustAnchor() && !isLeafStatement()
  }
  
  function stringifyJson(obj) {
    if (!obj) return ''
    return typeof obj === 'string' ? obj : JSON.stringify(obj, null, 2)
  }
  
  function handleJsonChange(field, value) {
    try {
      const parsed = JSON.parse(value)
      onEditChange(field, parsed)
    } catch (err) {
      onEditChange(field, value)
    }
  }
</script>

<div class="border border-gray-200 dark:border-gray-700 rounded-lg mb-0">
  <button
    on:click={onToggle}
    class="w-full flex items-center justify-between p-4 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors rounded-t-lg"
  >
    <div class="flex items-center gap-3">
      <span class="text-sm font-medium text-gray-500 dark:text-gray-400">#{index + 1}</span>
      <span class="px-3 py-1 text-xs font-medium rounded-full border {getTypeColor(statement.type)}">
        {getTypeLabel(statement.type)}
      </span>
      <span class="text-sm text-gray-900 dark:text-white font-medium">{formatEntityId(statement.subject)}</span>
    </div>
    
    <svg
      class="w-5 h-5 text-gray-500 dark:text-gray-400 transform transition-transform {isExpanded ? 'rotate-180' : ''}"
      fill="none"
      stroke="currentColor"
      viewBox="0 0 24 24"
    >
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
    </svg>
  </button>
  
  {#if isExpanded}
    <div class="border-t border-gray-200 dark:border-gray-700 p-4 bg-gray-50 dark:bg-gray-900/50 space-y-4">
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-gray-500 dark:text-gray-400">Issuer:</span>
          <p class="text-gray-900 dark:text-white font-mono text-xs mt-1 break-all">{statement.issuer}</p>
        </div>
        <div>
          <span class="text-gray-500 dark:text-gray-400">Subject:</span>
          <p class="text-gray-900 dark:text-white font-mono text-xs mt-1 break-all">{statement.subject}</p>
        </div>
        <div>
          <span class="text-gray-500 dark:text-gray-400">Issued At:</span>
          <p class="text-gray-900 dark:text-white">{formatTime(statement.iat)}</p>
        </div>
        <div>
          <span class="text-gray-500 dark:text-gray-400">Expires At:</span>
          <p class="text-gray-900 dark:text-white">{formatTime(statement.exp)}</p>
        </div>
      </div>
      
      <JWTViewer 
        header={statement.jwt_header} 
        payload={statement.jwt_payload} 
        rawJwt={statement.raw_jwt}
      />
      
      <!-- Inline Editing Section -->
      <div class="border-t border-gray-200 dark:border-gray-700 pt-4 mt-4">
        <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">Edit Payload Fields</h4>
        
        {#if canEditMetadata()}
          <div class="mb-4">
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
              Metadata (JSON) {isLeafStatement() ? '(Leaf)' : '(Direct Subordinate)'}
            </label>
            <textarea
              value={stringifyJson(editedStatement?.metadata)}
              on:change={(e) => handleJsonChange('metadata', e.target.value)}
              on:input={(e) => handleJsonChange('metadata', e.target.value)}
              rows="6"
              class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
            ></textarea>
          </div>
        {/if}
        
        {#if canEditMetadataPolicy()}
          <div class="mb-4">
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
              Metadata Policy (JSON)
            </label>
            <textarea
              value={stringifyJson(editedStatement?.metadata_policy)}
              on:change={(e) => handleJsonChange('metadata_policy', e.target.value)}
              on:input={(e) => handleJsonChange('metadata_policy', e.target.value)}
              rows="4"
              class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
            ></textarea>
          </div>
        {/if}
        
        {#if canEditConstraints()}
          <div class="mb-4">
            <label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">
              Constraints (JSON)
            </label>
            <textarea
              value={stringifyJson(editedStatement?.constraints)}
              on:change={(e) => handleJsonChange('constraints', e.target.value)}
              on:input={(e) => handleJsonChange('constraints', e.target.value)}
              rows="4"
              class="w-full px-2 py-1 text-sm border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white rounded font-mono focus:outline-none focus:ring-1 focus:ring-blue-500"
            ></textarea>
          </div>
        {/if}
        
        {#if isTrustAnchor()}
          <div class="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded p-3">
            <p class="text-xs text-yellow-800 dark:text-yellow-300 font-medium">Trust Anchor Entity Configuration</p>
            <p class="text-xs text-yellow-700 dark:text-yellow-400 mt-1">Trust Anchor statements cannot be edited. Only subordinate statements and the leaf entity configuration can be modified.</p>
          </div>
        {:else if !canEditMetadata() && !canEditMetadataPolicy() && !canEditConstraints()}
          <p class="text-xs text-gray-500 dark:text-gray-400 italic">No editable fields for this statement type</p>
        {/if}
      </div>
    </div>
  {/if}
</div>
