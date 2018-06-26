<template>
  <v-container v-if="loading" fluid fill-height>
    <v-layout align-center justify-center>
      <fingerprint-spinner :animation-duration="2000" :size="200" :color="'#9FC657'"></fingerprint-spinner>
    </v-layout>
  </v-container>
  <v-container v-else-if="task.isAbstract" class="pa-0" fluid fill-height>
    <v-layout column fill-height>
      <v-flex>
        <v-breadcrumbs v-if="project !== task.id">
          <v-breadcrumbs-item
            class="active"
            v-for="ancestor in ancestors"
            :key="ancestor.id"
            :disabled="false"
            :href="'/#/app/region/' + region + '/shard/' + shard + '/account/' + account + '/project/' + project + '/task/' + ancestor.id"
          >
            {{ ancestor.name }}
          </v-breadcrumbs-item>
          <v-breadcrumbs-item> <!--empty breadcrumb item as the last item is always disaabled, not suer if this is a bug or a feature but it's not helpful here -->
          </v-breadcrumbs-item>
        </v-breadcrumbs>
        <v-card>
          <v-card-title class="py-1"><h3>{{task.name}}</h3></v-card-title>
          <v-card-text class="py-1">{{task.description}}</v-card-text>
          <v-card-text class="py-1">
            Min: <h3>{{task.minimumRemainingTime}}</h3> &nbsp;
            Tot: <h3>{{task.totalRemainingTime}}</h3> &nbsp;
            Log: <h3>{{task.totalLoggedTime}}</h3>
          </v-card-text>
        </v-card>
      </v-flex>
      <v-data-table
        style="width: 100%!important; height: 100%!important"
        :headers="headers"
        :items="children"
        :loading="loading"
        hide-actions
        class="elevation-1"
      >
        <template slot="items" slot-scope="children">
          <tr @click="goToTask(children.item)" style="cursor: pointer">
            <td class="text-xs-left">{{ children.item.name }}</td>
            <td class="text-xs-left hidden-sm-and-down">{{ children.item.description? projects.item.description: 'none' }}</td>
            <td class="text-xs-left" style="width: 120px;">{{ children.item.minimumRemainingTime }}</td>
            <td class="text-xs-left" style="width: 120px;">{{ children.item.totalRemainingTime }}</td>
            <td class="text-xs-left" style="width: 120px;">{{ projects.item.totalLoggedTime }}</td>
          </tr>
        </template>
        <template slot="no-data">
          <v-alert v-if="!loading" :value="true" color="info" icon="error">
            No Child Tasks Yet <v-btn v-on:click="toggleCreateForm" color="primary">create</v-btn>
          </v-alert>
          <v-alert v-if="loading" :value="true" color="info" icon="error">
            Loading Data
          </v-alert>
        </template>
      </v-data-table>
      <v-btn v-if="children.length > 0" v-on:click="toggleCreateForm" fixed right bottom color="primary" fab>
        <v-icon>add</v-icon>
      </v-btn>
      <v-dialog
        v-model="createTaskDialog"
        fullscreen
        hide-overlay
        transition="dialog-bottom-transition"
        scrollable
      >
        <v-card tile>
          <v-toolbar card dark color="primary">
            <v-toolbar-title>Create Task</v-toolbar-title>
            <v-spacer></v-spacer>
            <v-btn icon dark @click.native="toggleCreateForm">
              <v-icon>close</v-icon>
            </v-btn>
          </v-toolbar>
          <v-card-text>
            <v-form ref="form" @keyup.native.enter="createTask">
              <v-text-field v-model="createTaskName" name="taskName" label="Name" type="text"></v-text-field>
              <v-text-field v-model="createTaskDescription" name="taskDescription" label="Description" type="text"></v-text-field>
              <v-switch
                :label="`Is Parallel`"
                v-model="createTaskIsParallel"></v-switch>
              <v-btn
                :loading="creating"
                :disabled="creating"
                color="secondary"
                @click="createTask"
              >
                Create
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-dialog>
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
      let params = router.currentRoute.params
      let data = {
        region: params.region,
        shard: params.shard,
        account: params.account,
        project: params.project,
        headers: [
          {text: 'Name', sortable: false, align: 'left', value: 'name'},
          {text: 'Description', class: 'hidden-sm-and-down', sortable: false, align: 'left', value: 'description'},
          {text: 'Min.', sortable: false, align: 'left', value: 'minimumRemainingTime'},
          {text: 'Tot.', sortable: false, align: 'left', value: 'totalRemainingTime'},
          {text: 'Log.', sortable: false, align: 'left', value: 'totalLoggedTime'}
        ],
        pagination: {
          descending: false
        },
        loading: true,
        createTaskDialog: false,
        createTaskName: '',
        createTaskDescription: null,
        createTaskIsParallel: true,
        creating: false,
        children: [],
        moreChildren: false,
        moreAncestors: false,
        task: null,
        ancestors: []
      }
      let mapi = api.newMGetApi(params.region)
      if (params.project === params.task) {
        mapi.v1.project.get(params.region, params.shard, params.account, params.project).then((project) => {
          data.task = project
          data.task.isAbstract = true
        })
      } else {
        mapi.v1.task.get(params.region, params.shard, params.account, params.project, params.task).then((task) => {
          data.task = task
        })
      }
      mapi.v1.task.getAncestors(params.region, params.shard, params.account, params.project, params.task, 100).then((res) => {
        data.ancestors = res.ancestors
        data.moreAncestors = res.more
      })
      mapi.v1.task.getChildren(params.region, params.shard, params.account, params.project, params.task, null, 100).then((res) => {
        data.children = res.children
        data.moreChildren = res.more
        data.loading = false
      })
      mapi.sendMGet()
      return data
    },
    mounted () {
      let htmlEl = document.querySelector('html')
      let self = this
      this.pageScrollListener = () => {
        if (self.moreChildren && htmlEl.clientHeight + htmlEl.scrollTop >= htmlEl.scrollHeight) {
          self.loadChildren()
        }
      }
      document.addEventListener('scroll', this.pageScrollListener)
    },
    beforeDestroy () {
      document.removeEventListener('scroll', this.pageScrollListener)
    },
    methods: {
      loadChildren () {
        if (!this.loading) {
          let params = router.currentRoute.params
          this.loading = true
          let after = null
          if (this.children.length > 0) {
            after = this.children[this.children.length - 1].id
          }
          api.v1.task.getChildren(params.region, params.shard, params.account, params.project, params.task, after, 100).then((res) => {
            this.loading = false
            res.children.forEach((child) => {
              this.children.push(child)
            })
            this.moreChildren = res.more
          })
        }
      },
      goToTask (task) {
        let params = router.currentRoute.params
        router.push('/app/region/' + params.region + '/shard/' + params.shard + /account/ + params.account + /project/ + params.project + /task/ + task.id)
      },
      toggleCreateForm () {
        if (!this.creating) {
          this.createTaskDialog = !this.createTaskDialog
          this.createTaskName = ''
          this.createTaskDescription = null
          this.createTaskIsParallel = true
        }
      },
      createTask () {
        this.creating = true
        let params = router.currentRoute.params
        let description = null
        if (this.createTaskDescription && this.createTaskDescription.length) {
          description = this.createTaskDescription
        }
        api.v1.task.create(params.region, params.shard, params.account, this.createTaskName, description).then((newProject) => {
          this.creating = false
          this.toggleCreateForm()
          let params = router.currentRoute.params
          router.push('/app/region/' + params.region + '/shard/' + params.shard + /account/ + params.account + /project/ + newProject.id + /task/ + newProject.id)
        })
      }
    }
  }
</script>

<style scoped lang="scss">
  h3, span {
    display: inline-flex;
  }
</style>
