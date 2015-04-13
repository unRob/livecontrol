import Live
import sys
import LiveUtils
from Log import Log
from Server import Server

class Interfaz:
    __module__ = __name__
    __doc__ = 'Clase principal'

    scene = 0
    track = 0
    slisten = {}
    clisten = {}
    # pplisten = {}
    cnlisten = {}
    cclisten = {}
    # wlisten = {}
    # llisten = {}

    def __init__(self, c_instance):
        self.c_instance = c_instance
        self._server = Server()
        self._setup_server()

        Log.info('Inicializado!')

        if self.song().visible_tracks_has_listener(self.refresh_state) != 1:
            Log.info('setting listener')
            self.song().add_visible_tracks_listener(self.refresh_state)


    # DE AHUEVO
    def connect_script_instances(self, instanciated_scripts):
        """
        Called by the Application as soon as all scripts are initialized.
        You can connect yourself to other running scripts here, as we do it
        connect the extension modules
        """
        return

    def is_extension(self):
        return False

    def request_rebuild_midi_map(self):
        """
        To be called from any components, as soon as their internal state changed in a
        way, that we do need to remap the mappings that are processed directly by the
        Live engine.
        Dont assume that the request will immediately result in a call to
        your build_midi_map function. For performance reasons this is only
        called once per GUI frame.
        """
        return

    def build_midi_map(self, midi_map_handle):
        self.refresh_state()

    def receive_midi(self, midi_bytes):
            return

    def can_lock_to_devices(self):
        return False

    def suggest_input_port(self):
        return ''

    def suggest_output_port(self):
        return ''

    def __handle_display_switch_ids(self, switch_id, value):
        pass
    # /DE AHUEVO


    def update_display(self):
        # logging.info('ud')
        try:
            self._server.read()
        except Exception, e:
            Log.info("chanfle")
            Log.info(e)
        return

    def refresh_state(self):
        Log.info('refresh_state')
        # self._add_scene_listeners()
        # self._add_tracks_listener()
        self.add_tempo_listener()
        self.add_clip_listeners()
        Log.info("stated")


    #------------
    #
    #------------
    #------------
    # Tempo
    #------------
    def add_tempo_listener(self):
        self.rem_tempo_listener()
        if self.song().tempo_has_listener(self.tempo_change) != 1:
            self.song().add_tempo_listener(self.tempo_change)

    def rem_tempo_listener(self):
        if self.song().tempo_has_listener(self.tempo_change) == 1:
            self.song().remove_tempo_listener(self.tempo_change)

    def tempo_change(self):
        tempo = self.song().tempo
        Log.info(tempo)
        self._server.send("tempo", str(tempo))


    def scene_change(self):
        selected_scene = self.song().view.selected_scene
        scenes = self.song().scenes
        index = 0
        selected_index = 0
        for scene in scenes:
            index = index + 1
            if scene == selected_scene:
                selected_index = index

        if selected_index != self.scene:
            self.scene = selected_index
            # self._server.send("scene:changed", selected_index)

    def track_change(self):
        selected_track = self.song().view.selected_track
        tracks = self.song().visible_tracks
        index = 0
        selected_index = 0
        for track in tracks:
            index = index + 1
            if track == selected_track:
                selected_index = index

        if selected_index != self.track:
            self.track = selected_index
            # self.server.send("track:changed", selected_index)

    def tracks_change(self):
        Log.info("Tracks change")


    #------------
    # Clips
    #------------
    def rem_clip_listeners(self):
        for slot in self.slisten:
            if slot != None:
                if slot.has_clip_has_listener(self.slisten[slot]) == 1:
                    slot.remove_has_clip_listener(self.slisten[slot])

        self.slisten = {}

        for clip in self.clisten:
            if clip != None:
                if clip.playing_status_has_listener(self.clisten[clip]) == 1:
                    clip.remove_playing_status_listener(self.clisten[clip])

        self.clisten = {}



    def add_clip_listeners(self):
        self.rem_clip_listeners()

        tracks = self.getslots()
        for track in range(len(tracks)):
            for clip in range(len(tracks[track])):
                c = tracks[track][clip]
                # if c.clip != None:
                self.add_cliplistener(c, track, clip)
                # Log.info("ClipLauncher: added clip listener tr: " + str(track) + " clip: " + str(clip));

                self.add_slotlistener(c, track, clip)

    def add_cliplistener(self, clip, tid, cid):
        cb = lambda :self.clip_changestate(clip, tid, cid)

        if self.clisten.has_key(clip) != 1:
            Log.info("Adding listener to "+str(tid)+"-"+str(cid))
            clip.add_playing_status_listener(cb)
            self.clisten[clip] = cb

    def add_slotlistener(self, slot, tid, cid):
        cb = lambda :self.slot_changestate(slot, tid, cid)

        if self.slisten.has_key(slot) != 1:
            slot.add_has_clip_listener(cb)
            self.slisten[slot] = cb


    # --------------
    # Clip Callbacks
    # --------------

    def slot_changestate(self, slot, tid, cid):
        tmptrack = LiveUtils.getTrack(tid)
        armed = tmptrack.arm and 1 or 0

        # Added new clip
        if slot.clip != None:
            self.add_cliplistener(slot.clip, tid, cid)

            playing = 1
            if slot.clip.is_playing == 1:
                playing = 2

            if slot.clip.is_triggered == 1:
                playing = 3

            length =  slot.clip.loop_end - slot.clip.loop_start

            self._server.send('track:info', {"track": tid, "armed":armed, "clip": cid, "status": str(playing), "length": str(length)})
        else:
            if self.clisten.has_key(slot.clip) == 1:
                slot.clip.remove_playing_status_listener(self.clisten[slot.clip])

            if self.cnlisten.has_key(slot.clip) == 1:
                slot.clip.remove_name_listener(self.cnlisten[slot.clip])

            if self.cclisten.has_key(slot.clip) == 1:
                slot.clip.remove_color_listener(self.cclisten[slot.clip])

            self._server.send('track:info', {"track": tid, "armed": armed, "clip": cid, "status": "0", "length": "0.0"})

        # Log.info("Slot changed" + str(self.clips[tid][cid]))

    def clip_changestate(self, clip, x, y):
        playing = 1

        if clip.is_playing == 1:
            playing = 2

        if clip.is_triggered == 1:
            playing = 3

        self._server.send('clip:state', {"track": x, "clip": y, "status": playing})
        Log.info("Clip changed x:" + str(x) + " y:" + str(y) + " status:" + str(playing))




    # Misc
    def getslots(self):
        tracks = self.song().visible_tracks
        clipSlots = []
        for track in tracks:
            clipSlots.append(track.clip_slots)
        return clipSlots




    # Server
    def _setup_server(self):
        self._server.on('record', self.start_recording)
        self._server.on('stop', self.stop_recording)
        self._server.on('play', self.play)


    # Inbound
    def start_recording(self, data):
        Log.info("Starting to record for "+data['track'])
        self.song().visible_tracks[int(data['track'])].clip_slots[int(data['clip'])].fire()

    def stop_recording(self, data):
        Log.info("Stopping clip "+data['track'])
        self.song().visible_tracks[int(data['track'])].clip_slots[int(data['clip'])].stop()

    def play(self, data):
        Log.info("Playing clip")
        self.song().visible_tracks[int(data['track'])].clip_slots[int(data['clip'])].fire()


    def disconnect(self):
        self._server.stop()
        self.rem_clip_listeners()
        # self._rem_tracks_listener()
        # self._rem_scene_listeners()
        Log.info('Chau!')


    def song(self):
        """returns a reference to the Live Song that we do interact with"""
        return self.c_instance.song()