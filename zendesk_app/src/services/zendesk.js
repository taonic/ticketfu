/**
 * Simple wrapper around ZAFClient
 * This allows us to:
 * 1. Mock it for testing
 * 2. Add consistent error handling
 * 3. Add additional functionality if needed
 */

let _client = null;

export default {
  /**
   * Initialize ZAF client
   * @returns {Object} ZAFClient instance
   */
  init() {
    if (!_client) {
      if (typeof ZAFClient === 'undefined') {
        throw new Error('ZAFClient is not defined. Make sure ZAF SDK is loaded.');
      }
      _client = ZAFClient.init();
    }
    return _client;
  },

  /**
   * Get the current client instance
   * @returns {Object} ZAFClient instance
   */
  getClient() {
    if (!_client) {
      return this.init();
    }
    return _client;
  }
};
