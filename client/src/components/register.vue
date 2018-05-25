<template>
  <div class="c-c">
    <div class="form">
      <input v-model="name" placeholder="username" type="text"/>
      <input v-model="email" placeholder="email" type="email"/>
      <input v-model="pwd" placeholder="password" type="password"/>
      <input v-model="displayName" placeholder="display name" type="text"/>
      <select v-model="selectedTheme">
        <option v-for="theme in themes" v-bind:value="theme.value">{{theme.text}}</option>
      </select>
      <button v-on:click="submit">submit</button>
    </div>
  </div>
</template>

<script>
  import api, {cnst} from '@/api'
  import router from '@/router'
  export default {
    name: 'register',
    data () {
      return {
        name: '',
        email: '',
        pwd: '',
        region: 'lcl',
        language: 'en',
        displayName: '',
        themes: [
          {text: 'Light', value: cnst.theme.light},
          {text: 'Dark', value: cnst.theme.dark},
          {text: 'Colorblind', value: cnst.theme.colorBlind}
        ],
        selectedTheme: cnst.theme.dark,
        err: null
      }
    },
    methods: {
      submit () {
        api.v1.centralAccount.register(this.name, this.email, this.pwd, this.region, this.language, this.displayName, this.selectedTheme).then(() => {
          router.push('/me')
        }).catch(() => {
          this.err = 'login failed'
        })
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
