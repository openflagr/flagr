import { createStore } from 'vuex';
import createPersistedState from 'vuex-persistedstate';
import { ABModeConstants, LatchModeConstants, MODES } from '../constants';

const store = createStore({
  state: {
    userDetails: null, // Define your shared state
    isAuth: false,
    mode: MODES.ABMode,
    terms: ABModeConstants
  },
  mutations: {
    setUserDetails(state, value) {
      state.userDetails = value; // Mutation to update the shared value
    },
    setAuth(state, value) {
      state.isAuth = value; // Mutation to update the shared value
    },
    setMode(state, value) {
      state.mode = value;
      if(value === MODES.ABMode) {
        state.terms = ABModeConstants;
      } else if(value === MODES.LatchMode) {
        state.terms = LatchModeConstants;
      }
    }
  },
  actions: {
    updateUserDetails({ commit }, value) {
      commit('setUserDetails', value); // Action to commit the mutation
    },
    updateAuth({ commit }, value) {
      commit('setAuth', value); // Action to commit the mutation
    },
    reset({ commit }) {
      commit('setUserDetails', null);
      commit('setAuth', false);
    },
    updateMode({ commit }, value) {
      commit('setMode', value);
    }
  },
  plugins: [
    createPersistedState({
      storage: window.localStorage, // Use localStorage to persist Vuex state
    }),
  ],
});

export default store;