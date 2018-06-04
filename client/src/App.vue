<template>
  <router-view></router-view>
</template>

<script>
  export default {
    name: 'app',
    data () {
      return {
        drawer: true
      }
    },
    props: {
      source: String
    }
  }
</script>

<style lang="scss">

</style>
