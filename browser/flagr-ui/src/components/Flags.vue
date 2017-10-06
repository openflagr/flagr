<template>
  <div>
    <spinner v-if="!loaded" />
    <div v-if="loaded">
      <ul >
        <li v-for="flag in flags">
          <router-link :to="{name: 'flag', params: {flagId: flag.id}}">
            [{{flag.id}}] {{ flag.description }}
          </router-link>
        </li>
      </ul>
      <form v-on:submit.prevent="createFlag">
        <p>
          <textarea
            placeholder="description"
            v-model="newFlag.description">  
          </textarea>
        </p>
        <input type="submit" value="Create Feature Flag" />
      </form>
    </div>
  </div>
</template>

<script>
import constants from '@/constants'
import fetchHelpers from '@/helpers/fetch'
import Spinner from '@/components/Spinner'

const {
  getJson,
  postJson
} = fetchHelpers

const {
  API_URL
} = constants

export default {
  name: 'flags',
  components: {
    spinner: Spinner
  },
  data () {
    return {
      loaded: false,
      flags: [],
      newFlag: {
        description: ''
      }
    }
  },
  created () {
    getJson(`${API_URL}/flags`)
      .then(flags => {
        this.loaded = true
        this.flags = flags
      })
  },
  methods: {
    createFlag () {
      postJson(`${API_URL}/flags`, this.newFlag)
        .then(flag => {
          this.flags.push(flag)
        })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h1, h2 {
  font-weight: normal;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  text-align: left;
  display: block;
  margin: 0 10px;
}

a {
  color: #42b983;
}
</style>
