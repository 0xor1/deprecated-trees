<template>
  <button v-on:click="click()">
    init
  </button>
</template>

<script>
  import api, {cnst} from '@/api'
  export default {
    name: 'init',
    data () {
      return {
        click: () => {
          console.log('clicked')
          let mgetApi = api.newMGetApi(cnst.regions.central)
          mgetApi.v1.centralAccount.getRegions().then((val) => {
            console.log(1)
            console.log(val)
          }).catch((error) => {
            console.log(1)
            console.log(error)
          })
          mgetApi.v1.centralAccount.getRegions().then((val) => {
            console.log(2)
            console.log(val)
          }).catch((error) => {
            console.log(2)
            console.log(error)
          })
          mgetApi.v1.centralAccount.getRegions().then((val) => {
            console.log(3)
            console.log(val)
          }).catch((error) => {
            console.log(3)
            console.log(error)
          })
          mgetApi.sendMGet()
        }
      }
    }
  }
</script>

<style scoped lang="scss">

</style>
