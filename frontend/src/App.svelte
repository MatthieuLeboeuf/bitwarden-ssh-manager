<script>
  import { GetEmail, IsLogged, Login, Sync, LaunchTerminal } from '../wailsjs/go/main/App.js';
  import { onMount } from 'svelte';

  let authStatus = -1;
  let authMessage = '';

  let password = '';
  let otpCode = null;
  let optRemember = false;

  let loading = false;

  let data = {hosts: [], folders: []};

  let currentFolder = {
    id: '',
    name: '',
  };

  async function auth() {
    authStatus = await Login(password, otpCode===null?0:otpCode, optRemember);
    switch (authStatus) {
      case 0:
        authMessage = '';
        await sync();
        break;
      case 2:
        authMessage = 'The totp code is invalid !';
        break;
      case 3:
        authMessage = 'An error has occurred !';
        break;
      case 4:
        authMessage = 'Rate Limit exceeded, please try again later.';
        break;
      case 5:
        authMessage = 'Password is incorrect. Try again';
        authStatus = -1;
        break;
    }
  }

  async function sync() {
    loading = true;
    data = JSON.parse(await Sync());
    loading = false;
  }

  async function authKey(e) {
    if (e.keyCode === 13) await auth();
  }

  async function setCurrentFolder(folder) {
    if (!folder) {
      currentFolder.id = '';
      return;
    }
    currentFolder.id = folder.id;
    currentFolder.name = folder.name;
  }

  onMount(async () => {
    if (await IsLogged()) {
      await sync();
      authStatus = 0;
      return;
    }
    authMessage = 'Logged in as '+ (await GetEmail());
  });
</script>

<main>
  {#if authStatus === authStatus}
    <nav class="navbar navbar-expand-lg bg-body-tertiary">
      <div class="container-fluid">
        <button
                class="navbar-toggler"
                type="button"
                data-bs-toggle="collapse"
                data-bs-target="#navbar"
                aria-controls="navbar"
                aria-expanded="false"
                aria-label="Toggle navigation"
        >
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbar">
          <div class="navbar-nav me-auto">
            <i class="bi bi-plus-square" style="font-size: 25px;"></i>
            <i class="bi bi-folder-plus mx-lg-3" style="font-size: 25px;"></i>
          </div>
          <i class="bi bi-arrow-clockwise" style="font-size: 25px;" on:click={sync} on:keypress={sync}></i>
        </div>
      </div>
    </nav>
  {/if}

  <div class="container mt-3">
    <p class="text-center">{authMessage}</p>
    {#if authStatus === -1 && !loading}
      <div class="row row-cols-lg-auto g-3 d-flex justify-content-center">
        <div class="col-12">
          <div class="input-group">
            <input type="password" bind:value={password} class="form-control" placeholder="Password" on:keyup={authKey}>
          </div>
        </div>

        <div class="col-12">
          <button class="btn btn-primary" on:click={auth}>Login</button>
        </div>
      </div>
    {/if}

    {#if (authStatus === 1 || authStatus === 2) && !loading}
      <div class="row row-cols-lg-auto g-3 d-flex justify-content-center">
        <div class="col-12">
          <div class="input-group">
            <input type="number" min="0" max="999999" bind:value={otpCode} class="form-control" placeholder="TOTP">
          </div>
        </div>

        <div class="col-12">
          <div class="input-group">
            <div class="form-check">
              <input class="form-check-input" type="checkbox" id="rememberOtp" bind:checked={optRemember}>
              <label class="form-check-label" for="rememberOtp">
                Remember
              </label>
            </div>
          </div>
        </div>

        <div class="col-12">
          <button class="btn btn-primary" on:click={auth}>Login</button>
        </div>
      </div>
    {/if}

    {#if loading}
      <div class="d-flex justify-content-center">
        <div class="spinner-border" role="status">
          <span class="visually-hidden">Loading...</span>
        </div>
      </div>
    {/if}

    {#if authStatus === 0}
      {#if currentFolder.id !== ''}
        <nav aria-label="breadcrumb">
          <ol class="breadcrumb">
            <li class="breadcrumb-item" on:click={() => setCurrentFolder()} on:keypress={() => setCurrentFolder()}>All hosts</li>
            <li class="breadcrumb-item active" aria-current="page">{currentFolder.name}</li>
          </ol>
        </nav>
      {/if}
      <div class="row row-cols-1 row-cols-xl-5 row-cols-lg-4 row-cols-md-3 row-cols-sm-2 g-3">
        {#if currentFolder.id === ''}
          {#each data.folders as folder}
            <div class="col">
              <div class="card h-100">
                <div class="card-body" on:click={() => setCurrentFolder(folder)} on:keypress={() => setCurrentFolder(folder)}>
                  <h5 class="card-title">{folder.name}</h5>
                  <p class="card-text">{folder.hosts} host{folder.hosts===1?'':'s'}</p>
                </div>
              </div>
            </div>
          {/each}
        {:else}
          {#each data.hosts as host}
            {#if host.folder === currentFolder.id}
              <div class="col">
                <div class="card h-100">
                  <div class="card-body" on:click={() => LaunchTerminal(JSON.stringify(host))} on:keypress={() => LaunchTerminal(JSON.stringify(host))}>
                    <h5 class="card-title">{host.name}</h5>
                    <p class="card-text">{host.username}, {host.password===""?"ssh key":"password"}</p>
                  </div>
                </div>
              </div>
            {/if}
          {/each}
        {/if}
      </div>
    {/if}
  </div>

  <!--<div class="input-box" id="input">
    <input autocomplete="off" bind:value={name} class="input" id="name" type="text" placeholder="Name"/>
    <input autocomplete="off" bind:value={address} class="input" id="address" type="text" style="margin-left: 20px" placeholder="Address"/>
    <button class="btn">Create</button>
  </div>-->
</main>
