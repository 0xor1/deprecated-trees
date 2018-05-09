<template>
  <div>
    init
  </div>
</template>

<script>
  import router from '@/router'
  export default {
    name: 'init',
    data () {
      return {
        click: () => {
          console.log('hello')
          console.log(router)
        }
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
