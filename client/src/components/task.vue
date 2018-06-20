<template>
  <v-container fluid fill-height class="pa-0">
    <v-layout v-if="task === null" align-center justify-center>
      <fingerprint-spinner :animation-duration="2000" :size="200" :color="'#9FC657'"></fingerprint-spinner>
    </v-layout>
    <v-layout v-else column>

    </v-layout>
  </v-container>
</template>

<script>
  import api from '@/api'
  import router from '@/router'
  import {FingerprintSpinner} from 'epic-spinners'
  export default {
    name: 'task',
    components: {FingerprintSpinner},
    data () {
      let data = {
        task: null,
        ancestors: null,
        children: null
      }
      let params = router.currentRoute.params
      let mapi = api.newMGetApi(params.region)
      if (params.project === params.task) {
        mapi.v1.project.get(params.region, params.shard, params.account, params.project).then((task) => {
          data.task = task
        })
      } else {
        mapi.v1.task.get(params.region, params.shard, params.account, params.project, params.task).then((task) => {
          data.task = task
        })
      }
      mapi.v1.task.getAncestors(params.region, params.shard, params.account, params.project, params.task, 100).then((ancestors) => {
        data.ancestors = ancestors
      })
      mapi.v1.task.getChildren(params.region, params.shard, params.account, params.project, params.task, null, 100).then((children) => {
        data.children = children
      })
      mapi.sendMGet()
      return data
    }
  }
</script>

<style scoped lang="scss">

</style>
