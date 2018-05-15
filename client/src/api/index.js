/**
 * IMPORTANT: this file should only be altered by backend api developers
 * **/

import axios from 'axios'

let newApi
newApi = (opts) => {
  let isMGetApi = opts.isMGetApi
  let mGetSending = false
  let mGetSent = false
  let awaitingMGetList = []

  let doReq = (axiosConfig) => {
    axiosConfig['X-Client'] = 'web'
    return axios(axiosConfig)
  }

  let get = (path, data) => {
    let url = path
    if (typeof data === 'object') {
      url = url + '?args=' + encodeURIComponent(JSON.stringify(data))
    }
    if (!isMGetApi || (mGetSending && !mGetSent)) {
      return doReq({
        method: 'get',
        url: url
      })
    } else if (isMGetApi && !mGetSending && !mGetSent) {
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

  let post = (path, data) => {
    doReq({
      method: 'post',
      url: path,
      data: data
    })
  }

  return {
    newMGetApi: () => {
      return newApi({isMGetApi: true})
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
        get('/api/mget', mgetObj).then((res) => {
          mGetSending = false
          mGetSent = true
          for (let i = 0, l = awaitingMGetList.length; i < l; i++) {
            let key = '' + i
            if (res.data[key].code === 200) {
              awaitingMGetList[i].resolve(res.data[key].body)
            } else {
              awaitingMGetList[i].reject(res.data[key].body)
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
    v1: {
      centralAccount: {
        getRegions: () => {
          return get('/api/v1/centralAccount/getRegions')
        },
        register: (name, email, pwd, region, language, displayName, theme) => {
          return post('/api/v1/centralAccount/register', {name, email, pwd, region, language, displayName, theme})
        },
        resendActivationEmail: (email) => {
          return post('/api/v1/centralAccount/resendActivationEmail', {email})
        },
        activate: (email, activationCode) => {
          return post('/api/v1/centralAccount/activate', {email, activationCode})
        },
        authenticate: (email, pwd) => {
          return post('/api/v1/centralAccount/authenticate', {email, pwd})
        },
        confirmNewEmail: (currentEmail, newEmail, confirmationCode) => {
          return post('/api/v1/centralAccount/confirmNewEmail', {currentEmail, newEmail, confirmationCode})
        },
        resetPwd: (email) => {
          return post('/api/v1/centralAccount/resetPwd', {email})
        },
        setNewPwdFromPwdReset: (newPwd, email, resetPwdCode) => {
          return post('/api/v1/centralAccount/setNewPwdFromPwdReset', {newPwd, email, resetPwdCode})
        },
        getAccount: (name) => {
          return get('/api/v1/centralAccount/getAccount', {name})
        },
        getAccounts: (accounts) => {
          return get('/api/v1/centralAccount/getAccounts', {accounts})
        },
        searchAccounts: (nameOrDisplayNameStartsWith) => {
          return get('/api/v1/centralAccount/searchAccounts', {nameOrDisplayNameStartsWith})
        },
        searchPersonalAccounts: (nameOrDisplayNameStartsWith) => {
          return get('/api/v1/centralAccount/namesearchPersonalAccounts', {nameOrDisplayNameStartsWith})
        },
        getMe: () => {
          return get('/api/v1/centralAccount/getMe')
        },
        setMyPwd: (oldPwd, newPwd) => {
          return post('/api/v1/centralAccount/setMyPwd', {oldPwd, newPwd})
        },
        setMyEmail: (newEmail) => {
          return post('/api/v1/centralAccount/setMyEmail', {newEmail})
        },
        resendMyNewEmailConfirmationEmail: () => {
          return post('/api/v1/centralAccount/resendMyNewEmailConfirmationEmail')
        },
        setAccountName: (account, newName) => {
          return post('/api/v1/centralAccount/setAccountName', {account, newName})
        },
        setAccountDisplayName: (account, newDisplayName) => {
          return post('/api/v1/centralAccount/setAccountDisplayName', {account, newDisplayName})
        },
        setAccountAvatar: (account, avatar) => {
          let data = new FormData()
          data.append('account', account)
          if (avatar) {
            data.append('avatar', avatar, '')
          }
          return post('/api/v1/centralAccount/setAccountAvatar', data)
        },
        migrateAccount: (account, newRegion) => {
          return post('/api/v1/centralAccount/migrateAccount', {account, newRegion})
        },
        createAccount: (name, region, displayName) => {
          return post('/api/v1/centralAccount/createAccount', {name, region, displayName})
        },
        getMyAccounts: (after, limit) => {
          return get('/api/v1/centralAccount/getMyAccounts', {after, limit})
        },
        deleteAccount: (account) => {
          return post('/api/v1/centralAccount/deleteAccount', {account})
        },
        addMembers: (account, newMembers) => {
          return post('/api/v1/centralAccount/addMembers', {account, newMembers})
        },
        removeMembers: (account, existingMembers) => {
          return post('/api/v1/centralAccount/removeMembers', {account, existingMembers})
        }
      },
      account: {
        setPublicProjectsEnabled: (account, publicProjectsEnabled) => {
          return post('/api/v1/account/setPublicProjectsEnabled', {account, publicProjectsEnabled})
        },
        getPublicProjectsEnabled: (account) => {
          return get('/api/v1/account/getPublicProjectsEnabled', {account})
        },
        setMemberRole: (account, member, role) => {
          return post('/api/v1/account/setMemberRole', {account, member, role})
        },
        getMembers: (account, role, nameContains, after, limit) => {
          return get('/api/v1/account/getMembers', {account, role, nameContains, after, limit})
        },
        getActivities: (account, item, member, occurredAfter, occurredBefore, limit) => {
          return get('/api/v1/account/getActivities', {account, item, member, occurredAfter, occurredBefore, limit})
        },
        getMe: (account) => {
          return get('/api/v1/account/getMe', {account})
        }
      },
      project: {
        createProject: (account, name, description, startOn, dueOn, isParallel, isPublic, members) => {
          return post('/api/v1/project/createProject', {account, name, description, startOn, dueOn, isParallel, isPublic, members})
        },
        setIsPublic: (account, project, isPublic) => {
          return post('/api/v1/project/setIsPublic', {account, project, isPublic})
        },
        setIsArchived: (account, project, isArchived) => {
          return post('/api/v1/project/setIsArchived', {account, project, isArchived})
        },
        getProject: (account, project) => {
          return get('/api/v1/project/getProject', {account, project})
        },
        getProjects: (account, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit) => {
          return get('/api/v1/project/getProjects', {account, nameContains, createdOnAfter, createdOnBefore, startOnAfter, startOnBefore, dueOnAfter, dueOnBefore, isArchived, sortBy, sortDir, after, limit})
        },
        deleteProject: (account, project) => {
          return post('/api/v1/project/deleteProject', {account, project})
        },
        addMembers: (account, project, members) => {
          return post('/api/v1/project/addMembers', {account, project, members})
        },
        setMemberRole: (account, project, member, role) => {
          return post('/api/v1/project/setMemberRole', {account, project, member, role})
        },
        removeMembers: (account, project, members) => {
          return post('/api/v1/project/removeMembers', {account, project, members})
        },
        getMembers: (account, project, role, nameOrDisplayNameContains, after, limit) => {
          return get('/api/v1/project/getMembers', {account, project, role, nameOrDisplayNameContains, after, limit})
        },
        getMe: (account, project) => {
          return get('/api/v1/project/getMembers', {account, project})
        },
        getActivities: (account, project, item, member, occurredAfter, occurredBefore, limit) => {
          return get('/api/v1/project/getActivities', {account, project, item, member, occurredAfter, occurredBefore, limit})
        }
      },
      task: {},
      timeLog: {}
    }
  }
}

export default newApi({false})
