import axios from 'axios'
import router from '@/router'

// important!! or axios will not send cookies to cors requests :( lost  3 days figuring this information out, shakes fist vigorously
axios.defaults.withCredentials = true

const config = {
  lcl: {
    central: 'http://lcl-api.project-trees.com',
    regions: {
      use: 'http://lcl-api.project-trees.com',
      usw: 'http://lcl-api.project-trees.com',
      euw: 'http://lcl-api.project-trees.com',
      asp: 'http://lcl-api.project-trees.com',
      aus: 'http://lcl-api.project-trees.com'
    }
  },
  dev: {
    central: 'https://dev-api.project-trees.com',
    regions: {
      use: 'https://dev-api.project-trees.com',
      usw: 'https://dev-api.project-trees.com',
      euw: 'https://dev-api.project-trees.com',
      asp: 'https://dev-api.project-trees.com',
      aus: 'https://dev-api.project-trees.com'
    }
  },
  stg: {
    central: 'https://stg-central-api.project-trees.com',
    regions: {
      use: 'https://stg-use-api.project-trees.com',
      usw: 'https://stg-usw-api.project-trees.com',
      euw: 'https://stg-euw-api.project-trees.com',
      asp: 'https://stg-asp-api.project-trees.com',
      aus: 'https://stg-aus-api.project-trees.com'
    }
  },
  pro: {
    central: 'https://pro-central-api.project-trees.com',
    regions: {
      use: 'https://pro-use-api.project-trees.com',
      usw: 'https://pro-usw-api.project-trees.com',
      euw: 'https://pro-euw-api.project-trees.com',
      asp: 'https://pro-asp-api.project-trees.com',
      aus: 'https://pro-aus-api.project-trees.com'
    }
  }
}

export const cnst = {
  regions: {
    central: 'central',
    use: 'use', // US East
    usw: 'usw', // US West
    euw: 'euw', // EU West
    asp: 'asp', // Asia Pacific
    aus: 'aus' // Australia
  },
  env: {
    lcl: 'lcl', // local
    dev: 'dev', // develop
    stg: 'stg', // staging
    pro: 'pro' // production
  },
  theme: {
    light: 0,
    dark: 1,
    colorBlind: 2
  },
  accountRole: {
    owner: 0,
    admin: 1,
    memberOfAllProjects: 2,
    memberOfOnlySpecificProjects: 3
  },
  projectRole: {
    admin: 0,
    writer: 1,
    reader: 2
  },
  sortBy: {
    name: 'name',
    displayName: 'displayname', // only used for users
    createdOn: 'createdon',
    // only used for projects
    startOn: 'starton',
    dueOn: 'dueon'
  }
}

let memCache = {}

let newApi
newApi = (opts) => {
  let isMGetApi = opts.isMGetApi
  let mGetApiRegion = opts.mGetApiRegion
  let mGetSending = false
  let mGetSent = false
  let awaitingMGetList = []
  let doReq = (axiosConfig) => {
    if (axiosConfig.data && typeof axiosConfig.data.shard === 'string') {
      axiosConfig.data.shard = parseInt(axiosConfig.data.shard, 10)
    }
    axiosConfig.headers = axiosConfig.headers || {}
    axiosConfig.headers['X-Client'] = 'web'
    return axios(axiosConfig).then((res) => {
      return res.data
    }).catch((res) => {
      if (res.response.status === 401) {
        router.push('/login')
      }
      throw res
    })
  }
  let getCentral = (path, data) => {
    return get(cnst.regions.central, path, data)
  }
  let postCentral = (path, data) => {
    return post(cnst.regions.central, path, data)
  }

  let get = (region, path, data) => {
    let url = path + '?region=' + region
    if (data) {
      if (typeof data.shard === 'string') {
        data.shard = parseInt(data.shard, 10)
      }
      url = url + '&args=' + encodeURIComponent(JSON.stringify(data))
    }
    if (!isMGetApi || (mGetSending && !mGetSent)) {
      return doReq({
        method: 'get',
        url: url
      })
    } else if (isMGetApi && !mGetSending && !mGetSent) {
      if (region !== mGetApiRegion) {
        throw new Error('invalid mget call, all get calls must be to the same region')
      }
      let awaitingMGetObj = {
        url: url,
        resolve: null,
        reject: null
      }
      awaitingMGetList.push(awaitingMGetObj)
      return new Promise((resolve, reject) => {
        awaitingMGetObj.resolve = resolve
        awaitingMGetObj.reject = reject
      })
    } else {
      throw new Error('invalid get call, use the default api object or a new mget instance from api.newMGetApi()')
    }
  }

  let post = (region, path, data) => {
    let url = path + '?region=' + region
    return doReq({
      method: 'post',
      url: url,
      data: data
    })
  }

  return {
    env: () => {
      return config.env
    },
    newMGetApi: (region) => {
      return newApi({isMGetApi: true, mGetApiRegion: region})
    },
    sendMGet: () => {
      if (!isMGetApi) {
        throw new Error('MGets must be made from the api instance returned from api.newMGetApi()')
      } else if (mGetSending || mGetSent) {
        throw new Error('each MGet must be started with a fresh api.newMGetApi(), once used that same instance cannot be reused')
      } else if (awaitingMGetList <= 1) {
        throw new Error('sending MGet requests should only be done with more than 1 request, otherwise just use a regular get')
      }
      mGetSending = true
      let asyncIndividualPromisesReady
      asyncIndividualPromisesReady = (resolve) => {
        for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
          if (awaitingMGetList[i].resolve === null) {
            setTimeout(asyncIndividualPromisesReady, 0, resolve)
          }
        }
        resolve()
      }
      new Promise(asyncIndividualPromisesReady).then(() => {
        let mgetObj = {}
        for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
          let key = '' + i
          mgetObj[key] = awaitingMGetList[i].url
        }
        get(mGetApiRegion, '/api/mget', mgetObj).then((res) => {
          mGetSending = false
          mGetSent = true
          for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
            let key = '' + i
            if (res[key].code === 200) {
              awaitingMGetList[i].resolve(res[key].body)
            } else {
              awaitingMGetList[i].reject(res[key])
            }
          }
        }).catch((error) => {
          mGetSending = false
          mGetSent = true
          for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
            awaitingMGetList[i].reject(error)
          }
        })
      })
    },
    logout: () => {
      memCache = {}
      return postCentral('/api/logout')
    },
    v1: {
      centralAccount: {
        register: (name, email, pwd, region, language, displayName, theme) => {
          return postCentral('/api/v1/centralAccount/register', {name, email, pwd, region, language, displayName, theme})
        },
        resendActivationEmail: (email) => {
          return postCentral('/api/v1/centralAccount/resendActivationEmail', {email})
        },
        activate: (email, activationCode) => {
          return postCentral('/api/v1/centralAccount/activate', {email, activationCode})
        },
        authenticate: (email, pwdTry) => {
          return postCentral('/api/v1/centralAccount/authenticate', {email, pwdTry}).then((res) => {
            memCache.me = res.me
            memCache[memCache.me.id] = memCache.me
            return res
          })
        },
        confirmNewEmail: (currentEmail, newEmail, confirmationCode) => {
          return postCentral('/api/v1/centralAccount/confirmNewEmail', {currentEmail, newEmail, confirmationCode})
        },
        resetPwd: (email) => {
          return postCentral('/api/v1/centralAccount/resetPwd', {email})
        },
        setNewPwdFromPwdReset: (newPwd, email, resetPwdCode) => {
          return postCentral('/api/v1/centralAccount/setNewPwdFromPwdReset', {newPwd, email, resetPwdCode})
        },
        getAccount: (name) => {
          return getCentral('/api/v1/centralAccount/getAccount', {name})
        },
        getAccounts: (accounts) => {
          return getCentral('/api/v1/centralAccount/getAccounts', {accounts})
        },
        searchAccounts: (nameOrDisplayNameStartsWith) => {
          return getCentral('/api/v1/centralAccount/searchAccounts', {nameOrDisplayNameStartsWith})
        },
        searchPersonalAccounts: (nameOrDisplayNameStartsWith) => {
          return getCentral('/api/v1/centralAccount/namesearchPersonalAccounts', {nameOrDisplayNameStartsWith})
        },
        getMe: () => {
          if (memCache.me) {
            return new Promise((resolve) => {
              resolve(memCache.me)
            })
          }
          return getCentral('/api/v1/centralAccount/getMe').then((res) => {
            memCache.me = res
            return res
          })
        },
        setMyPwd: (oldPwd, newPwd) => {
          return postCentral('/api/v1/centralAccount/setMyPwd', {oldPwd, newPwd})
        },
        setMyEmail: (newEmail) => {
          return postCentral('/api/v1/centralAccount/setMyEmail', {newEmail})
        },
        resendMyNewEmailConfirmationEmail: () => {
          return postCentral('/api/v1/centralAccount/resendMyNewEmailConfirmationEmail')
        },
        setAccountName: (account, newName) => {
          return postCentral('/api/v1/centralAccount/setAccountName', {account, newName})
        },
        setAccountDisplayName: (account, newDisplayName) => {
          return postCentral('/api/v1/centralAccount/setAccountDisplayName', {account, newDisplayName})
        },
        setAccountAvatar: (account, avatar) => {
          let data = new FormData()
          data.append('account', account)
          if (avatar) {
            data.append('avatar', avatar, '')
          }
          return postCentral('/api/v1/centralAccount/setAccountAvatar', data)
        },
        migrateAccount: (account, newRegion) => {
          return postCentral('/api/v1/centralAccount/migrateAccount', {account, newRegion})
        },
        createAccount: (name, region, displayName) => {
          return postCentral('/api/v1/centralAccount/createAccount', {name, region, displayName})
        },
        getMyAccounts: (after, limit) => {
          return getCentral('/api/v1/centralAccount/getMyAccounts', {after, limit})
        },
        deleteAccount: (account) => {
          return postCentral('/api/v1/centralAccount/deleteAccount', {account})
        },
        addMembers: (account, newMembers) => {
          return postCentral('/api/v1/centralAccount/addMembers', {account, newMembers})
        },
        removeMembers: (account, existingMembers) => {
          return postCentral('/api/v1/centralAccount/removeMembers', {account, existingMembers})
        }
      },
      account: {
        setPublicProjectsEnabled: (region, shard, account, publicProjectsEnabled) => {
          return post(region, '/api/v1/account/setPublicProjectsEnabled', {shard, account, publicProjectsEnabled})
        },
        getPublicProjectsEnabled: (region, shard, account) => {
          return get(region, '/api/v1/account/getPublicProjectsEnabled', {shard, account})
        },
        setMemberRole: (region, shard, account, member, role) => {
          return post(region, '/api/v1/account/setMemberRole', {shard, account, member, role})
        },
        getMembers: (region, shard, account, role, nameContains, after, limit) => {
          return get(region, '/api/v1/account/getMembers', {shard, account, role, nameContains, after, limit})
        },
        getActivities: (region, shard, account, item, member, occurredAfter, occurredBefore, limit) => {
          return get(region, '/api/v1/account/getActivities', {shard, account, item, member, occurredAfter, occurredBefore, limit})
        },
        getMe: (region, shard, account) => {
          return get(region, '/api/v1/account/getMe', {shard, account})
        }
      },
      project: {
        create: (region, shard, account, name, description, startOn, dueOn, isParallel, isPublic, members) => {
          return post(region, '/api/v1/project/create', {shard, account, name, description, startOn, dueOn, isParallel, isPublic, members})
        },
        setIsPublic: (region, shard, account, project, isPublic) => {
          return post(region, '/api/v1/project/setIsPublic', {shard, account, project, isPublic})
        },
        setIsArchived: (region, shard, account, project, isArchived) => {
          return post(region, '/api/v1/project/setIsArchived', {shard, account, project, isArchived})
        },
        get: (region, shard, account, project) => {
          return get(region, '/api/v1/project/get', {shard, account, project})
        },
        getSet: (region, shard, account, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit) => {
          return get(region, '/api/v1/project/getSet', {shard, account, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortAsc, after, limit})
        },
        delete: (region, shard, account, project) => {
          return post(region, '/api/v1/project/delete', {shard, account, project})
        },
        addMembers: (region, shard, account, project, members) => {
          return post(region, '/api/v1/project/addMembers', {shard, account, project, members})
        },
        setMemberRole: (region, shard, account, project, member, role) => {
          return post(region, '/api/v1/project/setMemberRole', {shard, account, project, member, role})
        },
        removeMembers: (region, shard, account, project, members) => {
          return post(region, '/api/v1/project/removeMembers', {shard, account, project, members})
        },
        getMembers: (region, shard, account, project, role, nameOrDisplayNameContains, after, limit) => {
          return get(region, '/api/v1/project/getMembers', {shard, account, project, role, nameOrDisplayNameContains, after, limit})
        },
        getMe: (region, shard, account, project) => {
          return get(region, '/api/v1/project/getMembers', {shard, account, project})
        },
        getActivities: (region, shard, account, project, item, member, occurredAfter, occurredBefore, limit) => {
          return get(region, '/api/v1/project/getActivities', {shard, account, project, item, member, occurredAfter, occurredBefore, limit})
        }
      },
      task: {
        create: (region, shard, account, project, parent, previousSibling, name, description, isAbstract, isParallel, member, remainingTime) => {
          return post(region, '/api/v1/task/create', {shard, account, project, parent, previousSibling, name, description, isAbstract, isParallel, member, remainingTime})
        },
        setName: (region, shard, account, project, task, name) => {
          return post(region, '/api/v1/task/setName', {shard, account, project, task, name})
        },
        setDescription: (region, shard, account, project, task, description) => {
          return post(region, '/api/v1/task/setDescription', {shard, account, project, task, description})
        },
        setIsParallel: (region, shard, account, project, task, isParallel) => {
          return post(region, '/api/v1/task/setDescription', {shard, account, project, task, isParallel})
        },
        setMember: (region, shard, account, project, task, member) => {
          return post(region, '/api/v1/task/setMember', {shard, account, project, task, member})
        },
        setRemainingTime: (region, shard, account, project, task, remainingTime) => {
          return post(region, '/api/v1/task/setremainingTime', {shard, account, project, task, remainingTime})
        },
        move: (region, shard, account, project, task, parent, nextSibling) => {
          return post(region, '/api/v1/task/move', {shard, account, project, task, parent, nextSibling})
        },
        delete: (region, shard, account, project, task) => {
          return post(region, '/api/v1/task/delete', {shard, account, project, task})
        },
        get: (region, shard, account, project, task) => {
          return get(region, '/api/v1/task/get', {shard, account, project, task})
        },
        getChildren: (region, shard, account, project, parent, fromSibling, limit) => {
          return get(region, '/api/v1/task/getChildren', {shard, account, project, parent, fromSibling, limit})
        },
        getAncestors: (region, shard, account, project, child, limit) => {
          return get(region, '/api/v1/task/getAncestors', {shard, account, project, child, limit})
        }
      },
      timeLog: {
        create: (region, shard, account, project, task, duration, note) => {
          return post(region, '/api/v1/task/create', {shard, account, project, task, duration, note})
        },
        createAndSetTimeRemaining: (region, shard, account, project, task, timeRemaining, duration, note) => {
          return post(region, '/api/v1/task/createAndSetTimeRemaining', {shard, account, project, task, timeRemaining, duration, note})
        },
        setDuration: (region, shard, account, project, timeLog, duration) => {
          return post(region, '/api/v1/task/setDuration', {shard, account, project, timeLog, duration})
        },
        setNote: (region, shard, account, project, timeLog, note) => {
          return post(region, '/api/v1/task/setNote', {shard, account, project, timeLog, note})
        },
        delete: (region, shard, account, project, timeLog) => {
          return post(region, '/api/v1/task/delete', {shard, account, project, timeLog})
        },
        get: (region, shard, account, project, task, member, timeLog, sortAsc, after, limit) => {
          return get(region, '/api/v1/task/get', {shard, account, project, task, member, timeLog, sortAsc, after, limit})
        }
      }
    }
  }
}

export default newApi({isMGetApi: false})
