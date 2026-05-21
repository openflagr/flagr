const API_URL = import.meta.env.VITE_API_URL || '/api/v1'
const rawEntityTypes = import.meta.env.VITE_FLAGR_UI_POSSIBLE_ENTITY_TYPES
const FLAGR_UI_POSSIBLE_ENTITY_TYPES = rawEntityTypes && rawEntityTypes !== 'null' ? rawEntityTypes : null

export default {
  API_URL,
  FLAGR_UI_POSSIBLE_ENTITY_TYPES
}
