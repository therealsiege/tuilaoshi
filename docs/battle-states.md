# Battle state chart

Generated from `internal/ui/rpg/states_test.go`, which also plays every
transition against the battle model. Regenerate with `make golden`.

```mermaid
stateDiagram-v2
    [*] --> Scene : open battle on a stage whose story scene is unseen
    [*] --> Intro : open battle (the open stage's demon or boss, else practice)
    Scene --> Intro : tab (scene marked seen)
    Scene --> [*] : esc (flee)
    Intro --> Fighting : enter / space
    Fighting --> PlayerAttack : right answer
    Fighting --> EnemyAttack : wrong answer
    Fighting --> EnemyAttack : time runs out (burning enemy)
    PlayerAttack --> Fighting : animation ends, enemy alive
    PlayerAttack --> Victory : animation ends, enemy HP 0 (saves; clears the stage)
    EnemyAttack --> Fighting : enter / space, player alive
    EnemyAttack --> Defeat : enter / space, player HP 0 (saves)
    Defeat --> Fighting : enter / space (rematch)
    Victory --> Intro : enter / space, more of the stage's foes to fight (HP carries)
    Victory --> [*] : enter / space, last foe beaten
    Fighting --> [*] : esc (flee, any state)
```
