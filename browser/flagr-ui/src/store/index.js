import { createStore } from 'vuex';
import createPersistedState from 'vuex-persistedstate';

const store = createStore({
  state: {
    userDetails: null, // Define your shared state
    isAuth: false
  },
  mutations: {
    setUserDetails(state, value) {
      console.log("setUserDetails", value)
      state.userDetails = value; // Mutation to update the shared value
    },
    setAuth(state, value) {
      console.log("setAuth", value)
      state.isAuth = value; // Mutation to update the shared value
    }
  },
  actions: {
    updateUserDetails({ commit }, value) {
      console.log("updateUserDetails", value)
      commit('setUserDetails', value); // Action to commit the mutation
    },
    updateAuth({ commit }, value) {
      console.log("updateAuth", value)
      commit('setAuth', value); // Action to commit the mutation
    },
    reset({ commit }) {
      commit('setUserDetails', null);
      commit('setAuth', false);
    }
  },
  plugins: [
    createPersistedState({
      storage: window.localStorage, // Use localStorage to persist Vuex state
    }),
  ],
});

export default store;