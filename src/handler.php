<?php

namespace CalProxy;


class handler {

    /** Associative array that maps building-ids to addresses */
    private $buildings, $courses;


    public function __construct() {
        // Load building addresses from file
        $this->buildings = json_decode(file_get_contents("buildings.json"), true);
        $this->courses = json_decode(file_get_contents("courses.json"), true);
    }

    /**
     * Parse the event and do the replacement and optimizations
     * @param $e Event a single ical event that should be cleaned up
     */
    public function cleanEvent(\ICal\Event &$e) {
        $event = new \Eluceo\iCal\Domain\Entity\Event(new \Eluceo\iCal\Domain\ValueObject\UniqueIdentifier($e->uid));

        //Strip added slashes by the parser
        $summary = stripcslashes($e->summary);
        $description = stripcslashes($e->description);
        $location = stripcslashes($e->location);

        //Remember the old title in the description
        $event->setDescription($summary . "\n" . $description);
        $event->setLocation(new \Eluceo\iCal\Domain\ValueObject\Location($location));

        //Remove the TAG and anything after e.g.: (IN0001) or [MA0001]
        $summary = preg_replace('/([(\[](?:(?:IN|MA|WI|WIB)\d+,?\s?)+[)\]]).+/', '', $summary);

        //remove location and teacher from language course title
        $summary = preg_replace('/(München|Garching|Weihenstephan).+/', '', $summary);

        //combine multiple spaces in summary into one
        $summary = preg_replace('/\s\s+/', ' ', $summary);

        //Do the replacement
        $summary = strtr($summary, $this->courses);

        //Remove some stuff which is not really needed
        $summary = str_replace(['Standardgruppe', 'PR, ', 'VO, ', 'FA, ', 'VI, ', 'TT, ', 'UE, ', 'SE, ','(Limited places) ', '(Online)'], '', $summary);

        //Clean up extra info for language course names
        if(preg_match('/(Spanisch|Französisch)\s(A|B|C)(1|2)((\.(1|2))|(\/(A|B|C)(1|2)))?(\s)/', $summary, $matches, PREG_OFFSET_CAPTURE) === 1){
            $summary = substr($summary, 0, $matches[10][1]);
        }

        //Try to make sense out of the location
        if (preg_match('/^(.*?),.*(\d{4})\.(?:\d\d|EG|UG|DG|Z\d|U\d)\.\d+/', $location, $matches) === 1) {
            $room = $matches[1]; // architect roomnumber (e.g. N1190)
            $b_id = $matches[2]; // 4-digit building-id (e.g. 0101)

            if (array_key_exists($b_id, $this->buildings)) {
                self::switchLocation($event, $location, $room . ", " . $this->buildings[$b_id]);
            }
        }

        //Check status
        switch ($e->status) {
            default:
            case 'CONFIRMED':
                $event->setStatus(\Eluceo\iCal\Domain\Enum\EventStatus::CONFIRMED());
                break;
            case 'CANCELLED':
                $event->setStatus(\Eluceo\iCal\Domain\Enum\EventStatus::CANCELLED());
                break;
            case 'TENTATIVE':
                $event->setStatus(\Eluceo\iCal\Domain\Enum\EventStatus::TENTATIVE());
                break;
        }

        //Add all fields
        $event
            ->touch(new \Eluceo\iCal\Domain\ValueObject\Timestamp(new \DateTime($e->dtstamp)))
            ->setSummary($summary)
            ->setOccurrence(new \Eluceo\iCal\Domain\ValueObject\TimeSpan(
                new \Eluceo\iCal\Domain\ValueObject\DateTime(new \DateTime($e->dtstart), true),
                new \Eluceo\iCal\Domain\ValueObject\DateTime(new \DateTime($e->dtend), true),
            ));

        return $event;
    }


    /**
     * Update the location field
     *
     * @param $e array element to be edited
     * @param $newLoc string new location that should be set to the element
     */
    public static function switchLocation(\Eluceo\iCal\Domain\Entity\Event &$e, $oldLocation, $newLoc) {
        $e->setDescription($oldLocation . "\n" . $e->getDescription());
        $e->setLocation(new \Eluceo\iCal\Domain\ValueObject\Location($newLoc, $oldLocation));
    }

    /**
     * Remove duplicate entries: events that happen at the same time in multiple locations
     * @param $events
     */
    public static function noDupes(array &$events) {
        //Sort them, first by starttime and then by title
        usort($events, function (\ICal\Event $a, \ICal\Event $b) {
            if (strtotime($a->dtstart) > strtotime($b->dtstart)) {
                return 1;
            } else if (strtotime($a->dtstart) < strtotime($b->dtstart)) {
                return -1;
            }
            //sort coinciding events according to their title
            return strcmp($a->summary, $b->summary);
        });

        //Find dupes
        $total = count($events);
        for ($i = 1; $i < $total; $i++) {
            //Check if start time, end time and title match then merge
            if ($events[$i - 1]->dtstart === $events[$i]->dtstart
                && $events[$i - 1]->dtend === $events[$i]->dtend
                && $events[$i - 1]->summary === $events[$i]->summary) {
                //if this and next event are the same, check whether the next is a livestream
                if (strpos($events[$i]->description, "Videoübertragung")) {
                    //Append the location to the next (same) element and switch the descriptions around
                    $events[$i]->location = $events[$i - 1]->location . "\n" . $events[$i]->location;
                    $events[$i]->description = $events[$i - 1]->description;
                    //Mark this element for removal
                    unset($events[$i - 1]);
                } else {
                    //Append the location to the next (same) element
                    $events[$i]->location .= "\n" . $events[$i - 1]->location;
                    //Mark this element for removal
                    unset($events[$i - 1]);
                }
            }
        }
    }
}
